package main

import (
	"machine"
	"time"

	"github.com/cswank/ir"
)

type (
	empty struct{}
)

const (
	led    = machine.LED
	btnPin = machine.GPIO12
	irPin  = machine.GPIO15
	ampPin = machine.GPIO16
	pwrPin = machine.GPIO17
)

var (
	btnPress bool
)

func main() {
	setupPins()

	btnInterrupt := func(p machine.Pin) {
		btnPress = true
	}

	btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)

	var i int
	var t2 time.Time
	t1 := time.Now()
	events := make([]time.Duration, 100)
	irEvents := func(p machine.Pin) {
		if i == 100 {
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
		if !(btnPress || (i >= ir.PayloadSize && time.Now().Sub(t1) > 100*time.Millisecond)) {
			continue
		}

		if btnPress {
			btnPin.SetInterrupt(0, nil)
			btnPress = false
			togglePower()
			btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)
		} else {
			irPin.SetInterrupt(0, nil)
			parseIR(events[:i])
			i = 0
			irPin.SetInterrupt(machine.PinToggle, irEvents)
		}
	}
}

func parseIR(events []time.Duration) {
	addr, cmd, err := ir.Command(events)
	if err != nil {
		blink(led, 5)
	} else {
		if addr == 0x35 && cmd == 0x40 {
			togglePower()
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

func togglePower() {
	ampPin.Set(!ampPin.Get())
	pwrPin.Set(!pwrPin.Get())
	blink(led, 2)
}

func setupPins() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ampPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	irPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	btnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()
}
