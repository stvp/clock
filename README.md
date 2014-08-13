clock
-----

A clock takes any number of strings and sends them at a somewhat consistent rate
on a channel. You can specify a tick interval and a cycle duration to control
how long it takes for the full set of keys to be cycled through. For example, if
you use an interval of 1 second and a cycle time of 1 minute and add 120 keys,
the clock will send a string on Channel 120 times per minute (2 per second on
average).

Any given key will always be placed at the same position on the clock as long as
the interval and cycle remain the same.

The clock can be stopped and started at any time.

```go
package main

import (
	"fmt"
	"github.com/stvp/clock"
	"time"
)

func main() {
	c, err := clock.New(100*time.Millisecond, time.Minute, 0)
	if err != nil {
		panic(err)
	}
	c.Add("neat")
	c.Add("dude")
	c.Add("rad")
	c.Start()

	for str := range c.Channel {
		fmt.Printf("Received: %s\n", str)
	}
}
```

