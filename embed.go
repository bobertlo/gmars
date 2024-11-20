package gmars

import (
	_ "embed"
)

var (
	//go:embed warriors/88/imp.red
	Imp_88_red []byte

	//go:embed warriors/94/imp.red
	Imp_94_red []byte

	//go:embed warriors/94/simpleshot.red
	SimpleShot_94_red []byte

	//go:embed warriors/94/bombspiral.red
	BombSpiral_94_red []byte
)
