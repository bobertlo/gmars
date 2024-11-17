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

var (
	ConfigKOTH88 = SimulatorConfig{
		Mode:       ICWS88,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
	}
	ConfigICWS88 = SimulatorConfig{
		Mode:       ICWS88,
		CoreSize:   8192,
		Processes:  8000,
		Cycles:     10000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     300,
		Distance:   100,
	}
	ConfigNOP94 = SimulatorConfig{
		Mode:       ICWS94,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
	}
	ConfigNopTiny = SimulatorConfig{
		Mode:       NOP94,
		CoreSize:   800,
		Processes:  800,
		Cycles:     8000,
		ReadLimit:  800,
		WriteLimit: 800,
		Length:     20,
		Distance:   20,
	}
	ConfigNopNano = SimulatorConfig{
		Mode:       NOP94,
		CoreSize:   80,
		Processes:  80,
		Cycles:     800,
		ReadLimit:  80,
		WriteLimit: 80,
		Length:     5,
		Distance:   5,
	}

	configPresets = map[string]SimulatorConfig{
		"88":      ConfigKOTH88,
		"icws":    ConfigICWS88,
		"nop94":   ConfigNOP94,
		"noptiny": ConfigNopTiny,
		"nopnano": ConfigNopNano,
	}
)

func PresetConfig(name string) (SimulatorConfig, error) {
	config, ok := configPresets[name]
	if !ok {
		return SimulatorConfig{}, fmt.Errorf("preset '%s' not found", name)
	}
	return config, nil
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
