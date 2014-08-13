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
type zilch interface{}

type Clock struct {
	Channel     chan string
	tick        time.Duration
	ticker      *time.Ticker
	slots       []map[string]zilch
	currentSlot int
}

func New(tick, cycle time.Duration) (clock *Clock, err error) {
	slotCount, err := ticksPerCycle(tick, cycle)
	if err != nil {
		return nil, err
	}

	slots := make([]map[string]zilch, slotCount)
	for i := range slots {
		slots[i] = map[string]zilch{}
	}

	clock = &Clock{
		Channel:     make(chan string),
		tick:        tick,
		slots:       slots,
		currentSlot: 0,
	}

	return clock, nil
}

func (c *Clock) Add(key string) error {
	index := c.slotIndex(key)
	_, found := c.slots[index][key]
	if found {
		return errors.New(fmt.Sprintf("%v already exists on the clock at position %d", key, index))
	}

	c.slots[index][key] = nil

	return nil
}

func (c *Clock) Remove(key string) error {
	index := c.slotIndex(key)
	_, found := c.slots[index][key]
	if !found {
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

func (c *Clock) doTick(index int) {
	for key, _ := range c.slots[index] {
		c.Channel <- key
	}
}

func (c *Clock) Start() {
	c.Stop()
	c.ticker = time.NewTicker(c.tick)
	go func() {
		for _ = range c.ticker.C {
			c.currentSlot = (c.currentSlot + 1) % len(c.slots)
			go func(index int) { c.doTick(index) }(c.currentSlot)
		}
	}()
}

func (c *Clock) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

// slotIndex returns the slot index for the given key.
func (c *Clock) slotIndex(key string) uint64 {
	return crc64.Checksum([]byte(key), hashTable) % uint64(len(c.slots))
}
