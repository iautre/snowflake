/**
 * @Author: a little
 * @Blog: https://coding.autre.cn
 * @Last Modified time: 2022-06-16 22:25:20
 */

// Twitter的Snowflake 算法
// 参考 hutool java版本https://gitee.com/dromara/hutool/blob/v5-master/hutool-core/src/main/java/cn/hutool/core/lang/Snowflake.java
package snowflake

import (
	"log"
	"sync"
	"time"
)

type Snowflake struct {
	mu sync.Mutex
	//10位的工作机器id
	workerId     int64 //5位工作id
	datacenterId int64 //5位数据id

	twepoch int64 //初始时间戳
	//上次时间戳，初始值为负数
	lastTimestamp int64
	sequence      int64 //12位序列
}

const (
	// 工作
	WORKER_ID_BITS      int64 = 5
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

	//默认值
	DEFAULT_LAST_TIMESTAMP int64 = -1
	DEFAULT_SEQUENCE       int64 = 0
	DEFAULT_TWEPOCH        int64 = 1155302795000
)

func NewSnowflake() *Snowflake {
	return &Snowflake{
		workerId:      1,
		datacenterId:  1,
		twepoch:       DEFAULT_TWEPOCH,
		lastTimestamp: DEFAULT_LAST_TIMESTAMP,
		sequence:      DEFAULT_SEQUENCE,
	}
}
func (s *Snowflake) WithTwepoch(twepoch int64) *Snowflake {
	s.twepoch = twepoch
	return s
}
func (s *Snowflake) WithWorkerId(workerId int64) *Snowflake {
	s.workerId = workerId
	return s
}
func (s *Snowflake) WithDatacenterId(datacenterId int64) *Snowflake {
	s.datacenterId = datacenterId
	return s
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
