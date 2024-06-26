package caesar

import (
	"sync"
)

type Clock struct {
	Timestamp uint64
	Mutex     *sync.Mutex
}

func NewClock() *Clock {
	return &Clock{
		Timestamp: 0,
		Mutex:     &sync.Mutex{},
	}
}

/* NewTimestamp returns a new timestamp for a new command and updates the global timestamp */
func (c *Clock) NewTimestamp() uint64 {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	newTS := c.Timestamp
	c.Timestamp += 1
	return newTS
}

/* SetTimestamp updates global timestamp if an observec command's timestamp is greater */
func (c *Clock) SetTimestamp(ts uint64) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if ts <	 c.Timestamp {
		return
	}
	c.Timestamp += 1
}
