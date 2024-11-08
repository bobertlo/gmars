# gMARS

[![Go
Reference](https://pkg.go.dev/badge/github.com/bobertlo/gmars/pkg/mars.svg)](https://pkg.go.dev/github.com/bobertlo/gmars/pkg/mars)

gMARS is an implementation of a Memory Array Redcode Simulator (MARS) written in
Go. The MARS simulator is used to play the game Core War, in which two or more
virus-like programs fight against each other in core memory. For more
information about Core War see:

- [corewar.co.uk](https://corewar.co.uk/): John Metcalf's Core War Site with
   tutorials, history, and links.
- [KOTH.org](http://www.koth.org/): A King of the Hill server with ongoing
   competitive matches, information, and links.
- [Koenigstuhl](https://asdflkj.net/COREWAR/koenigstuhl.html): An 'infinite
   hill' site that collects warriors and publishes their rankings and source
   code.

## Components

- `cmd/gmars` is a command line interface to run simulations
   and report results.
- `cmd/vmars` is a graphical simulator with interactive controls.
- The `pkg/mars` exports a public API for running MARS simulations.

## Implemented Features

- Load code (compiled assembly) warrior loading for ICWS'88 and '94 standards
   (without p-space)
- Simulation of two warrior battles
- Read/write limits (implemented, but not thoroughly tested)
- Hooks generating updates for visualization and analysis

## Planned Features

- P-Space support
- Parsing and linking of full '94 assembly spec (and pMARS compatibility)
- Interactive debugger
- GUI with interactive controls

## Testing Status / Known Bugs

> TL;DR: One warrior out of 1695 tested has divergent behavior from pMARS
> detected and I am searching for the issue.

To test for errors I used the 88 and 94nop hills from
[Koenigstuhl](https://asdflkj.net/COREWAR/koenigstuhl.html) and ran battles with
fixed starting positions to compare the output to pMARS and other
implementations.

## Results

The '88 hill has 658 warriors and all tested combinations / starting position
results matched.

The '94 hill has 1037 and all warriors had results matching pMARS except for
one. I am working on finding the bug there, but also working on new features at
the same time.
