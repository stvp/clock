package clock

import (
	"errors"
	"fmt"
	"time"
)

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
