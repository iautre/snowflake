// Package snowflake 实现 Twitter Snowflake 风格的 int64 ID（毫秒时间戳 + 数据中心 + 机器 + 序列）。
// 参考 hutool Java Snowflake 实现。
package snowflake

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// 常见错误（可用 errors.Is 判断）。
var (
	ErrClockMovedBackwards = errors.New("snowflake: system clock moved backwards")
	ErrTwepochInFuture     = errors.New("snowflake: twepoch is in the future")
	ErrInvalidWorkerID     = errors.New("snowflake: worker id out of range 0~31")
	ErrInvalidDatacenterID = errors.New("snowflake: datacenter id out of range 0~31")
	ErrTimeNotAdvancing    = errors.New("snowflake: time did not advance (possible clock stall)")
)

type Snowflake struct {
	mu            sync.Mutex
	workerId      int64
	datacenterId  int64
	twepoch       int64
	lastTimestamp int64
	sequence      int64
	err           error
}

const (
	WORKER_ID_BITS       int64 = 5
	DATA_CENTER_ID_BITS  int64 = 5
	MAX_WORKER_ID        int64 = -1 ^ (-1 << WORKER_ID_BITS)
	MAX_DATA_CENTER_ID   int64 = -1 ^ (-1 << DATA_CENTER_ID_BITS)
	SEQUENCE_BITS        int64 = 12
	SEQUENCE_MASK        int64 = -1 ^ (-1 << SEQUENCE_BITS)
	WORKER_ID_SHIFT      int64 = SEQUENCE_BITS
	DATA_CENTER_ID_SHIFT int64 = SEQUENCE_BITS + WORKER_ID_BITS
	TIMESTAMP_LEFT_SHIFT int64 = SEQUENCE_BITS + WORKER_ID_BITS + DATA_CENTER_ID_BITS

	defaultLastTimestamp int64 = -1
	defaultSequence      int64 = 0
	DefaultTwepoch       int64 = 1655647200000

	// DEFAULT_TWEPOCH 兼容旧版导出常量名，与 DefaultTwepoch 相同。
	DEFAULT_TWEPOCH = DefaultTwepoch

	maxTimeSpinIterations = 1_000_000
)

// NewSnowflake 创建生成器，worker/datacenter 均为 0。
func NewSnowflake() *Snowflake {
	return &Snowflake{
		workerId:      0,
		datacenterId:  0,
		twepoch:       DefaultTwepoch,
		lastTimestamp: defaultLastTimestamp,
		sequence:      defaultSequence,
	}
}

// NewSnowflakeFromEnv 从环境变量 SNOWFLAKE_WORKER_ID、SNOWFLAKE_DATACENTER_ID（0~31）读取；非法则忽略该项。
func NewSnowflakeFromEnv() *Snowflake {
	s := NewSnowflake()
	if v := os.Getenv("SNOWFLAKE_WORKER_ID"); v != "" {
		if w, e := strconv.ParseInt(v, 10, 64); e == nil && w >= 0 && w <= MAX_WORKER_ID {
			s.workerId = w
		}
	}
	if v := os.Getenv("SNOWFLAKE_DATACENTER_ID"); v != "" {
		if d, e := strconv.ParseInt(v, 10, 64); e == nil && d >= 0 && d <= MAX_DATA_CENTER_ID {
			s.datacenterId = d
		}
	}
	return s
}

// WithTwepoch 自定义纪元（毫秒），必须 <= 当前时间。
func (s *Snowflake) WithTwepoch(twepoch int64) *Snowflake {
	if s.err != nil {
		return s
	}
	if twepoch > s.currentTime() {
		s.err = fmt.Errorf("%w: twepoch=%d", ErrTwepochInFuture, twepoch)
		return s
	}
	s.twepoch = twepoch
	return s
}

func (s *Snowflake) WithWorkerId(workerId int64) *Snowflake {
	if s.err != nil {
		return s
	}
	if workerId < 0 || workerId > MAX_WORKER_ID {
		s.err = fmt.Errorf("%w: %d", ErrInvalidWorkerID, workerId)
		return s
	}
	s.workerId = workerId
	return s
}

func (s *Snowflake) WithDatacenterId(datacenterId int64) *Snowflake {
	if s.err != nil {
		return s
	}
	if datacenterId < 0 || datacenterId > MAX_DATA_CENTER_ID {
		s.err = fmt.Errorf("%w: %d", ErrInvalidDatacenterID, datacenterId)
		return s
	}
	s.datacenterId = datacenterId
	return s
}

// NextID 生成下一个 ID。
func (s *Snowflake) NextID() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.currentTime()
	if timestamp < s.lastTimestamp {
		return 0, fmt.Errorf("%w: last=%d current=%d", ErrClockMovedBackwards, s.lastTimestamp, timestamp)
	}

	if s.lastTimestamp == timestamp {
		s.sequence = (s.sequence + 1) & SEQUENCE_MASK
		if s.sequence == 0 {
			var err error
			timestamp, err = s.tilNextMillis()
			if err != nil {
				return 0, err
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = timestamp
	id := ((timestamp - s.twepoch) << TIMESTAMP_LEFT_SHIFT) |
		(s.datacenterId << DATA_CENTER_ID_SHIFT) |
		(s.workerId << WORKER_ID_SHIFT) |
		s.sequence
	return id, nil
}

// NextId 兼容旧 API：出错时 panic。新代码请用 NextID。
func (s *Snowflake) NextId() int64 {
	id, err := s.NextID()
	if err != nil {
		panic(err)
	}
	return id
}

// NextString 返回 NextID 的十进制字符串。
func (s *Snowflake) NextString() (string, error) {
	id, err := s.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// String 实现 fmt.Stringer；语义为再生成一个 ID 的字符串（易误解，优先用 NextString）。
func (s *Snowflake) String() string {
	str, err := s.NextString()
	if err != nil {
		panic(err)
	}
	return str
}

func (s *Snowflake) tilNextMillis() (int64, error) {
	var n int
	for {
		timestamp := s.currentTime()
		if timestamp > s.lastTimestamp {
			return timestamp, nil
		}
		n++
		if n >= maxTimeSpinIterations {
			return 0, ErrTimeNotAdvancing
		}
	}
}

func (s *Snowflake) currentTime() int64 {
	return time.Now().UnixMilli()
}

// ID 全局默认实例。
var ID *Snowflake

func init() {
	ID = NewSnowflakeFromEnv()
}

// NextID 使用全局默认实例。
func NextID() (int64, error) {
	return ID.NextID()
}

// NextId 使用全局默认实例，出错 panic。
func NextId() int64 {
	return ID.NextId()
}

// String 使用全局实例生成一个 ID 的十进制字符串（与历史版本兼容；出错时 panic，与 NextId 一致）。
func String() string {
	return strconv.FormatInt(NextId(), 10)
}
