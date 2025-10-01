.const base = $1000

// .for loop with scoped variable
.for (var loop_var = 0; loop_var < 4; loop_var++) {
    .byte loop_var
}

// This should generate an undefined symbol error
.byte loop_var

start:
    lda #$01
    rts