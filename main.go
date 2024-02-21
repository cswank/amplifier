package main

import (
	"machine"
	"time"

	"github.com/cswank/amplifier/internal/ir"
)

type (
	empty struct{}
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	irPin := machine.GPIO15
	irPin.Configure(machine.PinConfig{Mode: machine.PinInput})

	events := make([]time.Time, 200)
	var i int

	irEvents := func(p machine.Pin) {
		if i == 199 {
			return
		}

		events[i] = time.Now()
		i++
	}

	irPin.SetInterrupt(machine.PinToggle, irEvents)

	tk := time.NewTicker(250 * time.Millisecond)

	for {
		<-tk.C
		if i < 66 || time.Now().Sub(events[i-1]) < 100*time.Millisecond {
			continue
		}

		irPin.SetInterrupt(0, nil)
		parseIR(i, events, led)
		i = 0
		irPin.SetInterrupt(machine.PinToggle, irEvents)
	}
}

func blink(led machine.Pin, n int) {
	for i := 0; i < n; i++ {
		led.High()
		time.Sleep(100 * time.Millisecond)
		led.Low()
		time.Sleep(100 * time.Millisecond)
	}
}

func parseIR(i int, events []time.Time, led machine.Pin) {
	addr, cmd, err := ir.Command(events[:i])
	if err != nil {
		blink(led, 5)
	} else {
		if addr == 0x35 && cmd == 0x40 {
			blink(led, 2)
		} else {
			blink(led, 1)
		}
	}
}
