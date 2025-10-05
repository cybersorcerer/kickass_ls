// Test .enum directive parsing

.enum Colors {
    BLACK = 0,
    WHITE = 1,
    RED = 2,
    CYAN = 3,
    PURPLE = 4,
    GREEN = 5,
    BLUE = 6,
    YELLOW = 7
}

.enum Registers {
    BORDER = $d020,
    BACKGROUND = $d021,
    DUPLICATE = $d020  // Duplicate value - should warn
}

* = $0801
start:
    lda #Colors.RED
    sta Registers.BORDER
    lda #Colors.BLUE
    sta Registers.BACKGROUND
    rts
