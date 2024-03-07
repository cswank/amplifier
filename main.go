package main

import (
	"machine"
	"runtime/volatile"
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
	i        uint8
	t1, t2   time.Time
	index    volatile.Register8
	ch       chan empty
)

func main() {
	setup()

	for {
		<-ch

		switch shouldToggle() {
		case "button":
			btnPin.SetInterrupt(0, nil)
			togglePower()
			btnPress = false
			btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)
		case "ir":
			irPin.SetInterrupt(0, nil)
			parseIR()
			index.Set(0)
			irPin.SetInterrupt(machine.PinToggle, irInterrupt)
		default:
			if time.Now().Sub(t1) > 100*time.Millisecond {
				index.Set(0)
			}
		}
	}
}

func btnInterrupt(p machine.Pin) {
	btnPress = true
	select {
	case ch <- empty{}:
	default:
	}
}

func irInterrupt(p machine.Pin) {
	i = index.Get()
	if i == 100 {
		return
	}

	t2 = time.Now()
	irPulses[i] = t2.Sub(t1)
	t1 = t2
	i++
	index.Set(i)

	if i > ir.PayloadSize {
		select {
		case ch <- empty{}:
		default:
		}
	}
}

func shouldToggle() string {
	if btnPress {
		return "button"
	}

	time.Sleep(100 * time.Millisecond)
	if index.Get() >= ir.PayloadSize {
		return "ir"
	}

	return ""
}

func parseIR() {
	addr, cmd, err := ir.Command(irPulses[:index.Get()])
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
	pwrPin.Set(!st)
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
	irPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	btnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// both the power supply and amplifier are disabled by applying voltage to their respective on/off pins
	pwrPin.High()
	ampPin.High()

	onLEDPin.High()
	offLEDPin.Low()

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	btnPin.SetInterrupt(machine.PinToggle, btnInterrupt)
	irPin.SetInterrupt(machine.PinToggle, irInterrupt)

	ch = make(chan empty)
}
