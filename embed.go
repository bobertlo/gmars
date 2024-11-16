package gmars

import (
	_ "embed"
)

var (
	//go:embed warriors/88/imp.red
	imp_88_red []byte

	//go:embed warriors/94/imp.red
	imp_94_red []byte

	//go:embed warriors/94/simpleshot.red
	simpleshot_94_red []byte
)
