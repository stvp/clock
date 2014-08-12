package clock

import (
	"fmt"
	"github.com/stvp/assert"
	"testing"
	"time"
)

func TestTicksPerCycle(t *testing.T) {
	ticks, err := ticksPerCycle(time.Second, time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, uint64(60), ticks)

	ticks, err = ticksPerCycle(15*time.Second, time.Hour)
	assert.Nil(t, err)
	assert.Equal(t, uint64(240), ticks)

	_, err = ticksPerCycle(7*time.Second, time.Minute)
	assert.NotNil(t, err)

	_, err = ticksPerCycle(time.Minute, time.Second)
	assert.NotNil(t, err)
}

func TestNew(t *testing.T) {
	_, err := New(time.Millisecond, time.Second, 0)
	assert.NotNil(t, err)

	_, err = New(7*time.Millisecond, time.Second, 1)
	assert.NotNil(t, err)

	_, err = New(time.Minute, time.Second, 1)
	assert.NotNil(t, err)
}

func TestSlotIndex(t *testing.T) {
	clock, err := New(time.Millisecond, time.Minute, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), clock.slotIndex(""))
	assert.Equal(t, uint64(42695), clock.slotIndex("lol"))
	assert.Equal(t, uint64(44594), clock.slotIndex("omg"))

	// Check for a even distribution
	clock, err = New(6*time.Second, time.Minute, 1)
	assert.Nil(t, err)
	counts := make([]int, 10)
	for i := 0; i < 200; i++ {
		index := clock.slotIndex(fmt.Sprintf("%d:foobar:%d", i, i))
		counts[index] += 1
	}
	expected := []int{24, 21, 21, 15, 10, 20, 18, 21, 24, 26}
	assert.Equal(t, expected, counts, "should have a sane distribution")
}

func TestStartAndStop(t *testing.T) {
	clock, _ := New(10*time.Millisecond, 50*time.Millisecond, 4)
	received := []string{}
	go func() {
		for {
			received = append(received, <-clock.Channel)
		}
	}()
	clock.Add("foo") // slot 1
	clock.Add("biz") // slot 2
	clock.Add("baz") // slot 2
	clock.Add("fiz") // slot 4

	clock.Start()

	halfCycle := time.After(25 * time.Millisecond)
	fullCycle := time.After(50 * time.Millisecond)
	oneAndHalfCycles := time.After(75 * time.Millisecond)
	twoCycles := time.After(100 * time.Millisecond)

test:
	for {
		select {
		case <-halfCycle:
			assert.Equal(t, received, []string{"foo", "biz", "baz"})
		case <-fullCycle:
			assert.Equal(t, received, []string{"foo", "biz", "baz", "fiz"})
		case <-oneAndHalfCycles:
			assert.Equal(t, received, []string{"foo", "biz", "baz", "fiz", "foo", "biz", "baz"})
			clock.Stop()
		case <-twoCycles:
			assert.Equal(t, received, []string{"foo", "biz", "baz", "fiz", "foo", "biz", "baz"})
			break test
		}
	}
}

func TestBufferSize(t *testing.T) {
	clock, _ := New(time.Millisecond, 2*time.Millisecond, 3)
	clock.Add("foo")
	clock.Add("biz")
	clock.Add("baz")
	clock.doTick()
	clock.doTick()

	assert.Equal(t, 3, len(clock.Channel))
}

func TestKeys(t *testing.T) {
	clock, _ := New(10*time.Millisecond, 50*time.Millisecond, 1)
	clock.Add("foo") // slot 1
	clock.Add("biz") // slot 2
	clock.Add("baz") // slot 2

	expected := []string{"foo", "biz", "baz"}
	assert.Equal(t, expected, clock.Keys())
}
