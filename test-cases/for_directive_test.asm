.const base = $1000

// Simple .for loop
.for (var i = 0; i < 8; i++) {
    .byte i
}

// Nested .for loops with scope management
.for (var x = 0; x < 4; x++) {
    .for (var y = 0; y < 2; y++) {
        .word base + (x * 16) + y
    }
}

// .for loop with complex expressions
.for (var addr = base; addr < base + $100; addr += $10) {
    .byte <addr, >addr
}

// .for loop variable scope test
.const outside_var = 42
.for (var temp = 0; temp < 4; temp++) {
    .byte temp + outside_var
}
// temp should not be accessible here
// .byte temp  // This should generate undefined symbol error

start:
    lda #$01
    rts