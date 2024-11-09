# gMARS

[![Go
Reference](https://pkg.go.dev/badge/github.com/bobertlo/gmars/pkg/mars.svg)](https://pkg.go.dev/github.com/bobertlo/gmars/pkg/mars)
[![Go Report
Card](https://goreportcard.com/badge/github.com/bobertlo/gmars)](https://goreportcard.com/report/github.com/bobertlo/gmars)

gMARS is an implementation of a Core War Memory Array Redcode Simulator (MARS)
written in Go.

![vmars screenshot](screenshot.png)

In the the game of Core War, two virus-like programs written in assembly fight
against each other in the memory of a Simulator where they are able to write and
execute new code and modify eachother. For more information about Core War see:

- [corewar.co.uk](https://corewar.co.uk/): John Metcalf's Core War Site with
   tutorials, history, and links.
- [KOTH.org](http://www.koth.org/): A King of the Hill server with ongoing
   competitive matches, information, and links.
- [Koenigstuhl](https://asdflkj.net/COREWAR/koenigstuhl.html): An 'infinite
   hill' site that collects warriors and publishes their rankings and source
   code.

## Running

Currently only two warrior matches are supported. The simulator setting
arguments are specified the same for the GUI and CLI simulators:

```
Usage of ./gmars:
  -8    Enforce ICWS'88 rules
  -F int
        fixed position of warrior #2
  -c int
        Cycles until tie (default 80000)
  -debug
        Dump verbose debug information
  -l int
        Max. warrior length (default 100)
  -p int
        Max. Processes (default 8000)
  -r int
        Rounds to play (default 1)
  -s int
        Size of core (default 8000)
```

## Install

You can use `go install` to build and install the executables from the latest
release:

```
# vmars - Visual MARS simulator
go install github.com/bobertlo/gmars/cmd/vmars@latest

# gmars - command line MARS
go install github.com/bobertlo/gmars/cmd/gmars@latest
```

## Components

- `cmd/gmars` is a command line interface to run simulations and report results.
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
