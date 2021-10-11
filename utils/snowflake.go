package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	nodeBits        = 10
	stepBits        = 12
	nodeMax         = -1 ^ (-1 << nodeBits)
	stepMask  int64 = -1 ^ (-1 << stepBits)
	timeShift uint8 = nodeBits + stepBits
	nodeShift uint8 = stepBits
)

type ID int64

var Epoch int64 = 1288834974657

type Snowflake struct {
	mutex sync.Mutex

	time int64
	node int64
	step int64
}

func New(node int64) (*Snowflake, error) {
	if node < 0 || node > nodeMax {
		return nil, fmt.Errorf("node number must be between %d and %d", 0, nodeMax)
	}

	return &Snowflake{
		time: 0,
		node: node,
		step: 0,
	}, nil
}

func (sf *Snowflake) Next() ID {
	sf.mutex.Lock()

	now := ms()
	switch {
	case now < sf.time:
		now = wait(now, sf.time)
	case now == sf.time:
		sf.step = (sf.step + 1) & stepMask
		if sf.step == 0 {
			now = wait(now, sf.time)
		}
	case now > sf.time:
		sf.step = 0
	}
	sf.time = now

	id := (now-Epoch)<<timeShift |
		sf.node<<nodeShift |
		sf.step

	sf.mutex.Unlock()

	return ID(id)
}

func ms() int64 {
	return time.Now().UnixNano() / 1000000
}

func wait(now int64, target int64) int64 {
	for now <= target {
		now = ms()
	}
	return now
}

func (id ID) Int64() int64 {
	return int64(id)
}

func (id ID) Uint64() uint64 {
	return uint64(id)
}

func (id ID) String() string {
	return fmt.Sprintf("%d", id.Int64())
}
