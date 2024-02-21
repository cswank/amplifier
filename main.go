package main

import (
	"machine"
	"time"

	necir "github.com/cswank/amplifier/internal/ir"
)

type (
	empty struct{}
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	irPin := machine.GPIO15
	irPin.Configure(machine.PinConfig{Mode: machine.PinInput})

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	ir := necir.New()
	ch := make(chan time.Time)
	go wait(irPin, ch)

	var addr, cmd uint8
	var err error
	var ts time.Time
	for {
		ts = <-ch
		ir.Add(ts)
		if !ir.Ready() {
			continue
		}

		addr, cmd, err = ir.Result()
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
}

func blink(led machine.Pin, n int) {
	for i := 0; i < n; i++ {
		led.High()
		time.Sleep(100 * time.Millisecond)
		led.Low()
		time.Sleep(100 * time.Millisecond)
	}
}

func wait(gpio machine.Pin, ch chan time.Time) {
	gpio.SetInterrupt(machine.PinToggle, func(p machine.Pin) {
		ch <- time.Now()
	})
}

type timer struct {
	t    *time.Timer
	recv bool
}

func newTimer(d time.Duration) *timer {
	return &timer{t: time.NewTimer(d)}
}

func (t *timer) reset(d time.Duration) {
	if !t.t.Stop() {
		if !t.recv {
			<-t.t.C
		}
	}
}
