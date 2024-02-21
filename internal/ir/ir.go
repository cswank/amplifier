package ir

import (
	"fmt"
	"sync"
	"time"
)

const (
	tolerance  = 100 * time.Microsecond
	start      = 9 * time.Millisecond
	startSpace = 4500 * time.Microsecond
	bitStart   = 562500 * time.Nanosecond
	bitOne     = 1687500 * time.Nanosecond

	initial   = state(0)
	lead      = state(1)
	leadSpace = state(2)
	data      = state(3)
	ready     = state(4)
)

type (
	state int

	IR struct {
		lock  *sync.Mutex
		state state
		buf   []time.Duration
		ch    chan time.Time
	}
)

func New() *IR {
	i := &IR{
		lock: &sync.Mutex{},
		ch:   make(chan time.Time),
		buf:  make([]time.Duration, 64),
	}
	go i.add()
	return i
}

func (i *IR) Add(t time.Time) {
	i.ch <- t
}

func (i *IR) add() {
	var t1, t2 time.Time
	var d time.Duration
	var j int
	for {
		t2 = <-i.ch
		i.lock.Lock()
		d = t2.Sub(t1)
		t1 = t2

		switch i.state {
		case initial:
			i.state++
		case lead:
			if closeTo(d, start) {
				i.state++
			}
		case leadSpace:
			if closeTo(d, startSpace) {
				i.state++
			} else {
				i.state = initial
			}
		case data:
			i.buf[j] = d
			j++
			if j == 64 {
				i.state++
				j = 0
			}
		}
		i.lock.Unlock()
	}
}

func (i *IR) Ready() bool {
	i.lock.Lock()
	b := i.state == ready
	i.lock.Unlock()
	return b
}

func (i *IR) Result() (addr, cmd uint8, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.state = initial

	addr, err = i.parse(0)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid addr %s", err)
	}

	iAddr, err := i.parse(1)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid iaddr %s", err)
	}

	cmd, err = i.parse(2)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid cmd %s", err)
	}

	iCmd, err := i.parse(3)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid icmd %s", err)
	}

	if iAddr != addr^0xff {
		return 0, 0, fmt.Errorf("invalid address 0x%x, inverse 0x%x", addr, iAddr)
	}

	if iCmd != cmd^0xff {
		return 0, 0, fmt.Errorf("invalid command 0x%x, inverse 0x%x", cmd, iCmd)
	}

	return addr, cmd, nil
}

func (i *IR) parse(pos int) (uint8, error) {
	var mask uint8
	var val uint8

	for j, d := range i.buf[pos*16 : (pos*16 + 16)] {
		if j%2 == 0 {
			if !closeTo(d, bitStart) {
				return 0, fmt.Errorf("start bit %d", j/2)
			}
		} else {
			if closeTo(d, bitStart) {
				mask = 0
			} else if closeTo(d, bitOne) {
				mask = 1 << (j / 2)
			} else {
				return 0, fmt.Errorf("data bit %d", j/2)
			}
			val ^= mask
		}
	}

	return val, nil
}

func closeTo(d time.Duration, val time.Duration) bool {
	//fmt.Printf("%t %s (%s - %s)\n", d >= val-tolerance && d <= val+tolerance, d, val-tolerance, val+tolerance)
	return d >= val-tolerance && d <= val+tolerance
}
