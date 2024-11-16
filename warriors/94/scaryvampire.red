;redcode
;name Scary Vampire
;author Robert Lowry
;strategy vampire

        org     vamp+1
        step    equ 6192

vamp    add     inc,        fang
        mov     fang,       @fang
        jmz.f   vamp,       *fang
        mov     fang,       *fang
        jmz.f   vamp,       trap
        jmp     clear

fang    jmp     @step,      trap-step
inc     dat     step,       -step

gate    dat     bomb,       100

dbmb    dat     bomb-gate,  9
bomb    spl     #dbmb-gate, 11
clear   mov     *gate,      >gate
        mov     *gate,      >gate
        djn.f   clear,      {-250


trap    spl     #0,         {0
        spl     {0,         }0
        jmn.a   trap+1,     trap
