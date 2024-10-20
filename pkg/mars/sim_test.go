package mars

func makeSim94() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, false)
}

func makeSim88() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, true)
}
