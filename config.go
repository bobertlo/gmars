package gmars

import (
	"fmt"

	"math/rand"
)

type SimulatorConfig struct {
	Mode       SimulatorMode
	CoreSize   Address
	Processes  Address
	Cycles     Address
	ReadLimit  Address
	WriteLimit Address
	Length     Address
	Distance   Address
	Fixed      Address
}

func ConfigKOTH88() SimulatorConfig {
	return SimulatorConfig{
		Mode:       ICWS88,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
		Fixed:      0,
	}
}

func ConfigNOP94() SimulatorConfig {
	return SimulatorConfig{
		Mode:       ICWS94,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
		Fixed:      0,
	}
}

func NewQuickConfig(mode SimulatorMode, coreSize, processes, cycles, length, fixed Address) SimulatorConfig {
	out := SimulatorConfig{
		Mode:       mode,
		CoreSize:   coreSize,
		Processes:  processes,
		Cycles:     cycles,
		ReadLimit:  coreSize,
		WriteLimit: coreSize,
		Length:     length,
		Distance:   length,
		Fixed:      fixed,
	}
	return out
}

func (c SimulatorConfig) Validate() error {
	if c.CoreSize < 3 {
		return fmt.Errorf("the minimum core size is 3")
	}

	if c.Processes < 1 {
		return fmt.Errorf("invalid process limit")
	}

	if c.ReadLimit < 1 {
		return fmt.Errorf("invalid read limit")
	}
	if c.WriteLimit < 1 {
		return fmt.Errorf("invalid read limit")
	}

	if c.Cycles < 1 {
		return fmt.Errorf("invalid cycle count")
	}

	if c.Length > c.CoreSize {
		return fmt.Errorf("invalid warrior length")
	}

	if c.Length+c.Distance > c.CoreSize {
		return fmt.Errorf("invalid distance")
	}

	if c.Fixed != 0 && c.Fixed+c.Length+1 > c.CoreSize {
		return fmt.Errorf("invalid fixed starting point")
	}

	return nil
}

func (c SimulatorConfig) GetW2Start() Address {
	w2start := c.Fixed
	if w2start == 0 {
		minStart := 2 * c.Length
		maxStart := c.CoreSize - c.Length - 1
		startRange := maxStart - minStart
		w2start = Address(rand.Intn(int(startRange)+1) + int(minStart))
	}

	return w2start
}
