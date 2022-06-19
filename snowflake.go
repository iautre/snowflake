// Twitter的Snowflake 算法
// 参考 hutool java版本https://gitee.com/dromara/hutool/blob/v5-master/hutool-core/src/main/java/cn/hutool/core/lang/Snowflake.java
package snowflake

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type Snowflake struct {
	mu sync.Mutex
	//10位的工作机器id
	workerId      int64 //5位工作id
	datacenterId  int64 //5位数据id
	twepoch       int64 //初始时间戳
	lastTimestamp int64 //上次时间戳，初始值为负数
	sequence      int64 //12位序列
}

const (
	// 工作id
	WORKER_ID_BITS int64 = 5
	//数据中心id
	DATA_CENTER_ID_BITS int64 = 5
	//最大值
	MAX_WORKER_ID      int64 = -1 ^ (-1 << WORKER_ID_BITS)
	MAX_DATA_CENTER_ID int64 = -1 ^ (-1 << DATA_CENTER_ID_BITS)
	//序列号id长度
	SEQUENCE_BITS int64 = 12
	//序列号最大值
	SEQUENCE_MASK int64 = -1 ^ (-1 << SEQUENCE_BITS)
	//工作id需要左移的位数，12位
	WORKER_ID_SHIFT int64 = SEQUENCE_BITS
	//数据id需要左移位数 12+5=17位
	DATA_CENTER_ID_SHIFT int64 = SEQUENCE_BITS + WORKER_ID_BITS
	//时间戳需要左移位数 12+5+5=22位
	TIMESTAMP_LEFT_SHIFT int64 = SEQUENCE_BITS + WORKER_ID_BITS + DATA_CENTER_ID_BITS

	// 默认值
	DEFAULT_LAST_TIMESTAMP int64 = -1
	// 默认序列开始值
	DEFAULT_SEQUENCE int64 = 0
	// 默认时间戳 2022-06-19 22:00:00
	DEFAULT_TWEPOCH int64 = 1655647200000
)

func NewSnowflake() *Snowflake {
	return &Snowflake{
		workerId:      0,
		datacenterId:  0,
		twepoch:       DEFAULT_TWEPOCH,
		lastTimestamp: DEFAULT_LAST_TIMESTAMP,
		sequence:      DEFAULT_SEQUENCE,
	}
}
func (s *Snowflake) WithTwepoch(twepoch int64) *Snowflake {
	if twepoch > s.currentTime() {
		log.Fatal("twepoch should < time now")
	}
	if twepoch < DEFAULT_TWEPOCH {
		log.Fatal("twepoch should > 2022-06-19 22:00:00")
	}
	s.twepoch = twepoch
	return s
}
func (s *Snowflake) WithWorkerId(workerId int64) *Snowflake {
	if workerId > MAX_WORKER_ID || workerId < 0 {
		log.Fatal("workerId should 0~31")
	}
	s.workerId = workerId
	return s
}
func (s *Snowflake) WithDatacenterId(datacenterId int64) *Snowflake {
	if datacenterId > MAX_DATA_CENTER_ID || datacenterId < 0 {
		log.Fatal("datacenterId should 0~31")
	}
	s.datacenterId = datacenterId
	return s
}
func (s *Snowflake) String() string {
	return strconv.FormatInt(s.NextId(), 10)
}
func (s *Snowflake) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := s.currentTime()
	if timestamp < s.lastTimestamp {
		// panic(errors.New("系统时间异常"))
		log.Fatal("system time error")
	}
	if s.lastTimestamp == timestamp {
		s.sequence = (s.sequence + 1) & SEQUENCE_MASK
		if s.sequence == 0 {
			timestamp = s.tilNextMillis()
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = timestamp
	id := ((timestamp - s.twepoch) << TIMESTAMP_LEFT_SHIFT) |
		(s.datacenterId << DATA_CENTER_ID_SHIFT) |
		(s.workerId << WORKER_ID_SHIFT) |
		s.sequence
	return id
}

// 等待下一秒
func (s *Snowflake) tilNextMillis() int64 {
	timestamp := s.currentTime()
	for timestamp <= s.lastTimestamp {
		timestamp = s.currentTime()
	}
	return timestamp
}

// 当前时间戳
func (s *Snowflake) currentTime() int64 {
	return time.Now().UnixNano() / 1e6
}

var ID *Snowflake

func init() {
	ID = NewSnowflake()
}

func NextId() int64 {
	return ID.NextId()
}
func String() string {
	return strconv.FormatInt(ID.NextId(), 10)
}
