# gMARS

gMARS is an implementation of a Memory Array Redcode Simulator (MARS) written in
Go. The MARS simulator is used to play the game Core War, in which two or more
virus-like programs fight against each other in core memory. For more
information about Core War see:

- [corewar.co.uk](https://corewar.co.uk/): John Metcalf's Core War Site with
   tutorials, history, and links.
- [KOTH.org](http://www.koth.org/): A King of the Hill server with ongoing
   competitive matches, information, and links.
- [Koenigstuhl](https://asdflkj.net/COREWAR/koenigstuhl.html): An infinite hill that collects most warriors with known source code and calculates their rankings.

## Project Status

This project is still a work in progress, but the core simulator is functional,
tested, and mostly bug free. gMARS does include a CLI client, but is implemented
as a library, with generic reporting hooks to be used by other software to
implement Graphical User Interface and/or analysis software.

While I do plan to optimize the code as much as possible, there is no goal to
compete with the many optimized C/C++ implementations on raw performance. This
project is more focused on readability and a stable API for creating interactive
applications for educational and analytical purposes.

### Implemented Features

- Load code (compiled assembly) warrior loading for ICWS'88 and '94 standards
   (without p-space)
- Simulation of two warrior battles
- Read/write limits (implemented, but not thoroughly tested)
- Hooks generating updates for visualization and analysis

### Planned Features

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

### Results

The '88 hill has 658 warriors and all tested combinations / starting position
results matched.

The '94 hill has 1037 and all warriors had results matching pMARS except for
one. I am working on finding the bug there, but also working on new features at
the same time.
