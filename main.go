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

	var i int
	var t2 time.Time
	t1 := time.Now()
	events := make([]time.Duration, 200)
	irEvents := func(p machine.Pin) {
		if i == 199 {
			return
		}

		t2 = time.Now()
		events[i] = t2.Sub(t1)
		t1 = t2
		i++
	}

	irPin.SetInterrupt(machine.PinToggle, irEvents)

	tk := time.NewTicker(250 * time.Millisecond)

	for {
		<-tk.C
		if i < ir.PayloadSize || time.Now().Sub(t1) < 100*time.Millisecond {
			continue
		}

		irPin.SetInterrupt(0, nil)
		parseIR(events[:i], led)
		i = 0
		irPin.SetInterrupt(machine.PinToggle, irEvents)
	}
}

func parseIR(events []time.Duration, led machine.Pin) {
	addr, cmd, err := ir.Command(events)
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

func blink(led machine.Pin, n int) {
	for i := 0; i < n; i++ {
		led.High()
		time.Sleep(100 * time.Millisecond)
		led.Low()
		time.Sleep(100 * time.Millisecond)
	}
}
