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
	led       = machine.LED
	btnPin    = machine.GPIO12
	irPin     = machine.GPIO15
	ampPin    = machine.GPIO20
	pwrPin    = machine.GPIO21
	onLEDPin  = machine.GPIO17
	offLEDPin = machine.GPIO16
)

var (
	irPulses [100]time.Duration
	btnPress bool
	i        int
	t1, t2   time.Time
	tk       *time.Ticker
)

func main() {
	setup()

	for {
		<-tk.C

		switch shouldToggle() {
		case "button":
			btnPin.SetInterrupt(0, nil)
			togglePower()
			btnPress = false
			btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)
		case "ir":
			irPin.SetInterrupt(0, nil)
			parseIR()
			i = 0
			irPin.SetInterrupt(machine.PinToggle, irInterrupt)
		}
	}
}

func btnInterrupt(p machine.Pin) {
	btnPress = true
}

func irInterrupt(p machine.Pin) {
	if i == 100 {
		return
	}

	t2 = time.Now()
	irPulses[i] = t2.Sub(t1)
	t1 = t2
	i++
}

func shouldToggle() string {
	if btnPress {
		return "button"
	}

	if i >= ir.PayloadSize && time.Now().Sub(t1) > 100*time.Millisecond {
		return "ir"
	}

	return ""
}

func parseIR() {
	addr, cmd, err := ir.Command(irPulses[:i])
	if err != nil {
		blink(led, 5)
		return
	}

	if addr == 0x35 && cmd == 0x40 {
		togglePower()
	} else {
		blink(led, 1) // not the button we're looking for
	}
}

func blink(led machine.Pin, n int) {
	for i := 0; i < n; i++ {
		led.High()
		time.Sleep(500 * time.Millisecond)
		led.Low()
		time.Sleep(500 * time.Millisecond)
	}
}

func togglePower() {
	st := ampPin.Get()
	ampPin.Set(!st)
	pwrPin.Set(!pwrPin.Get())
	onLEDPin.Set(!st)
	offLEDPin.Set(st)
	blink(led, 2)
}

func setup() {
	t1 = time.Now()

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ampPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	onLEDPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	offLEDPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	onLEDPin.Low()
	offLEDPin.High()
	pwrPin.Low()
	ampPin.Low()

	irPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	btnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	tk = time.NewTicker(250 * time.Millisecond)

	btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)
	irPin.SetInterrupt(machine.PinToggle, irInterrupt)
}
