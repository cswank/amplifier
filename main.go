package main

import (
	"encoding/json"
	"machine"
	"os"
	"time"
)

type (
	empty struct{}
	event struct {
		Duration string `json:"duration"`
		State    bool   `json:"state"`
	}
)

func main() {

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	irCh := make(chan bool)
	ir := machine.GPIO15
	ir.Configure(machine.PinConfig{Mode: machine.PinInput})
	go wait(ir, irCh)

	led.High()
	time.Sleep(250 * time.Millisecond)
	led.Low()

	t1 := time.Now()

	tk := time.NewTicker(5 * time.Second)

	enc := json.NewEncoder(os.Stdout)
	events := make([]event, 100)
	var i int
	var t2 time.Time
	for {
		select {
		case <-tk.C:
			enc.Encode(events[:i])
			i = 0
		case st := <-irCh:
			if i > 99 {
				continue
			}

			t2 = time.Now()
			events[i] = event{Duration: t2.Sub(t1).String(), State: st}
			t1 = t2
			i++
		}
	}
}

func wait(gpio machine.Pin, ch chan bool) {
	gpio.SetInterrupt(machine.PinToggle, func(p machine.Pin) {
		ch <- gpio.Get()
	})
}
