// Package mars implements interfaces for MARS simulators and helper
// functions to define configuration and load warrior files.
//
// The Simulator interface provides a MARS implemenation for running
// simulations, and the ReportingSimulator interface adds the Addreporter
// method to inject Reporter interfaces to recieve callbacks to report state
// changes in the simulation.
//
// RedCode files are first loaded as WarriorData{} structs, holding the data
// needed to create a Warrior inside a Simulator. These can safely be reused
// to create multiple Simulators concurrently.
//
// The SimulatorConfig struct is provided to define configuration when
// creating Simulator instances, and compiling warriors.
package mars
