;redcode-94b
;name Paper Haze
;author Robert Lowry
;strategy quickbomb -> paper
;assert CORESIZE==8000

a for 20
	mov <100+(350*a), 266+(350*a)
rof

step    equ 1092

        spl    1
        spl    1
paper   spl    step,     {src
        mov    }src,     }paper
src     mov    *paper+4, }paper
        jmz.f  @paper+1, *src
