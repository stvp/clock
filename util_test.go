package clock

import (
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
