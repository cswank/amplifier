package ir

import (
	"fmt"
	"time"
)

const (
	tolerance  = 100 * time.Microsecond
	start      = 9 * time.Millisecond
	startSpace = 4500 * time.Microsecond
	bitStart   = 562500 * time.Nanosecond
	bitOne     = 1687500 * time.Nanosecond
)

type (
	cmd struct {
		addr  uint8
		iAddr uint8
		cmd   uint8
		iCmd  uint8
	}
)

func Command(times []time.Time) (addr, cmd uint8, err error) {
	if len(times) < 67 {
		return 0, 0, fmt.Errorf("not enough data, must have a length of at least 67")
	}
	var t1 time.Time
	for i, t2 := range times {
		if i == 0 {
			t1 = t2
			continue
		}

		d := t2.Sub(t1)
		t1 = t2

		if closeTo(d, start) {
			return command(times[i:])
		}
	}

	return 0, 0, fmt.Errorf("unable to find beginning of a valid command")
}

func command(times []time.Time) (addr, cmd uint8, err error) {
	t1 := times[0]
	t2 := times[1]
	d := t2.Sub(t1)
	if !closeTo(d, startSpace) {
		return 0, 0, fmt.Errorf("invalid commad, expected a %s space", startSpace)
	}

	addr, err = parse("addr", times[1:])
	if err != nil {
		return 0, 0, err
	}

	iAddr, err := parse("iAddr", times[17:])
	if err != nil {
		return 0, 0, err
	}

	cmd, err = parse("cmd", times[33:])
	if err != nil {
		return 0, 0, err
	}

	iCmd, err := parse("iCmd", times[49:])
	if err != nil {
		return 0, 0, err
	}

	if iAddr != addr^0xff {
		return 0, 0, fmt.Errorf("invalid address %d, inverse %d", addr, iAddr)
	}

	if iCmd != cmd^0xff {
		return 0, 0, fmt.Errorf("invalid command %d, inverse %d", cmd, iCmd)
	}

	return addr, cmd, nil
}

func parse(typ string, times []time.Time) (val uint8, err error) {
	var t1 time.Time
	var i int
	for j, t2 := range times[:17] {
		if j == 0 {
			t1 = t2
			continue
		}

		d := t2.Sub(t1)
		t1 = t2

		if i%2 == 0 {
			if !closeTo(d, bitStart) {
				return 0, fmt.Errorf("invalid %s", typ)
			}
		} else {
			var mask uint8
			if closeTo(d, bitStart) {
				// add a zero
			} else if closeTo(d, bitOne) {
				mask = 1 << (i / 2)
			} else {
				return 0, fmt.Errorf("invalid %s", typ)
			}
			val ^= mask
		}
		i++
	}

	return val, nil
}

func closeTo(d time.Duration, val time.Duration) bool {
	return d >= val-tolerance && d <= val+tolerance
}
