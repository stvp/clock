package clock

import (
	"errors"
	"fmt"
	"hash/crc64"
	"time"
)

var (
	// ISO polynomial results in uneven distributions and is considered weak for
	// hashing: http://www0.cs.ucl.ac.uk/staff/d.jones/crcnote.pdf The ECMA
	// polynomial is much more evenly distributed.
	hashTable = crc64.MakeTable(crc64.ECMA)
)

// A zilch takes up 0 bytes of space.
type zilch struct{}

// ticksPerCycle returns the number of ticks that will occur given a tick
// duration and a total cycle duration.
func ticksPerCycle(tick, cycle time.Duration) (ticks uint64, err error) {
	tickMs := uint64(tick / time.Millisecond)
	cycleMs := uint64(cycle / time.Millisecond)
	if cycleMs < tickMs {
		return 0, errors.New(fmt.Sprintf("The cycle time (%v) is less than the tick time (%v)", cycle, tick))
	}
	if cycleMs%tickMs != 0 {
		return 0, errors.New(fmt.Sprintf("The cycle time (%v) is not evenly divisible by the tick time (%v)", cycle, tick))
	}
	return cycleMs / tickMs, nil
}

type Clock struct {
	Channel     chan string
	tick        time.Duration
	ticker      *time.Ticker
	slots       []map[string]zilch
	currentSlot int
}

func New(tick, cycle time.Duration, bufferSize uint) (clock *Clock, err error) {
	if bufferSize < 1 {
		return nil, fmt.Errorf("Channel buffer size must be greater than 0 (got %d)", bufferSize)
	}
	slotCount, err := ticksPerCycle(tick, cycle)
	if err != nil {
		return nil, err
	}
	slots := make([]map[string]zilch, slotCount)
	for i := range slots {
		slots[i] = map[string]zilch{}
	}
	clock = &Clock{
		Channel:     make(chan string, bufferSize),
		tick:        tick,
		slots:       slots,
		currentSlot: 0,
	}
	return clock, nil
}

// slotIndex returns the slot index for the given key.
func (c *Clock) slotIndex(key string) uint64 {
	return crc64.Checksum([]byte(key), hashTable) % uint64(len(c.slots))
}

func (c *Clock) Add(key string) error {
	index := c.slotIndex(key)
	if _, found := c.slots[index][key]; found {
		return errors.New(fmt.Sprintf("%v already exists on the clock at position %d", key, index))
	}
	c.slots[index][key] = zilch{}
	return nil
}

func (c *Clock) Remove(key string) error {
	index := c.slotIndex(key)
	if _, found := c.slots[index][key]; !found {
		return errors.New(fmt.Sprintf("%v couldn't be found in the clock", key))
	}
	delete(c.slots[index], key)
	return nil
}

func (c *Clock) Keys() (keys []string) {
	keys = []string{}
	for _, slot := range c.slots {
		for key, _ := range slot {
			keys = append(keys, key)
		}
	}
	return keys
}

func (c *Clock) doTick() {
	c.currentSlot = (c.currentSlot + 1) % len(c.slots)
	for key, _ := range c.slots[c.currentSlot] {
		if len(c.Channel) < cap(c.Channel) {
			c.Channel <- key
		}
	}
}

func (c *Clock) Start() {
	c.Stop()
	c.ticker = time.NewTicker(c.tick)
	go func() {
		for {
			<-c.ticker.C
			c.doTick()
		}
	}()
}

func (c *Clock) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	// TODO: Stop the goroutine started in Start().
}
