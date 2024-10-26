package mars

import (
	"testing"
)

func makeSim94() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, false)
}

func makeSim88() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, true)
}

func TestImp(t *testing.T) {
	tests := []redcodeTest{
		{
			input: []string{
				"mov.i #0, $1",
			},
			output: []string{
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
			},
			turns: 16,
			pq:    []Address{0},
		},
	}
	runTests(t, "dwarf", tests)
}

func TestDwarf(t *testing.T) {
	tests := []redcodeTest{
		{
			input: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #0",
			},
			output: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #12",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #4",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #8",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #12",
			},
			turns: 9,
			pq:    []Address{0},
		},
	}
	runTests(t, "dwarf", tests)
}
