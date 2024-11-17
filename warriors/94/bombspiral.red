;redcode-94b
;name bomb spiral
;author Robert Lowry
;strategy stone and imp launcher

        spl istart
        jmp bstart

        sptr   equ -1333
        step   equ 953
        time   equ 3382

        spl    #0,       0
        spl      0,       0
stone   mov    bomb,     hit+step*time
hit     add    #-step,   stone
        djn.f  stone,    <5555
bomb    dat    >-1,      {1

go      dat    #0,       #sptr
bstart  mov    {-1,      <-1
        mov    {-2,      <-2
        mov    {-3,      <-3
        mov    {-4,      <-4
        mov    {-5,      <-5
        mov    {-6,      <-6
        jmp @go

for 75
dat 0, 0
rof

        istep  equ 1143           ; (CORESIZE+1)/3

istart  spl    #0,         >prime
prime   mov    imp,        imp
        add.a  #istep+1,   launch
launch  jmp    imp-istep-1

imp     mov.i  #0,         istep
