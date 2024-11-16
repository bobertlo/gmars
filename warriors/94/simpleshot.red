;redcode
;name Simple Shot
;author Robert Lowry
;strategy decoy -> one shot

        org     decoy

        first   equ 51
        gap     equ 19
        step    equ 404

        dpos    equ decoy+4000

scan    add     inc,        gate
gate    sne     first+gap,  }first
        djn.f   scan,       {338

        jmp     clear

bptr    dat    bomb,       9
bomb    spl    #2700,       11
clear   mov    *bptr,      >gate
        mov    *bptr,      >gate
        djn.f  clear,      }bomb

inc     dat     step,       step

decoy   nop    >dpos,    }dpos+1
        mov    {dpos+2,  <dpos+4
        mov    {dpos+5,  <dpos+7
        mov    {dpos+8,  <dpos+10
        mov    {dpos+11, <dpos+13
        djn.f  scan,     {dpos+15
