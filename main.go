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

	irCh := make(chan empty)
	ir := machine.GPIO15
	ir.Configure(machine.PinConfig{Mode: machine.PinInput})
	go wait(ir, irCh)

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	tk := time.NewTicker(250 * time.Millisecond)
	events := make([]time.Time, 200)
	var addr, cmd uint8
	var err error
	var i int
	for {
		select {
		case <-tk.C:
			if i == 0 || time.Now().Sub(events[i-1]) < 100*time.Millisecond {
				continue
			}

			addr, cmd, err = necir.Command(events[:i])
			if err != nil {
				blink(led, 5)
			} else {
				if addr == 0x35 && cmd == 0x40 {
					blink(led, 2)
				} else {
					blink(led, 1)
				}
			}
			i = 0
		case <-irCh:
			if i == 199 {
				continue
			}

			events[i] = time.Now()
			i++
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

func wait(gpio machine.Pin, ch chan empty) {
	gpio.SetInterrupt(machine.PinToggle, func(p machine.Pin) {
		ch <- empty{}
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
	t.t.Reset(d)
}
