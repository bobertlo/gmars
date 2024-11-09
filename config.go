package gmars

import "fmt"

type SimulatorConfig struct {
	Mode       SimulatorMode
	CoreSize   Address
	Processes  Address
	Cycles     Address
	ReadLimit  Address
	WriteLimit Address
	Length     Address
	Distance   Address
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
	}
}

func NewQuickConfig(mode SimulatorMode, coreSize, processes, cycles, length Address) SimulatorConfig {
	out := SimulatorConfig{
		Mode:       mode,
		CoreSize:   coreSize,
		Processes:  processes,
		Cycles:     cycles,
		ReadLimit:  coreSize,
		WriteLimit: coreSize,
		Length:     length,
		Distance:   length,
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

	return nil
}
