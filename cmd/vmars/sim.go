package main

func (g *Game) runCycle() {
	if g.finished {
		return
	}

	count := g.sim.WarriorCount()
	living := g.sim.WarriorLivingCount()
	if ((count > 1 && living > 1) || living > 0) && g.sim.CycleCount() < g.sim.MaxCycles() {
		g.sim.RunCycle()
	} else {
		g.finished = true
	}
}
