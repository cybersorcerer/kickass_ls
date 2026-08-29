// Test .enum directive parsing
//
// .enum takes no name (manual 4.3). Wrapping each group in a namespace keeps
// the qualified access below working (manual 9.4).

.namespace Colors {
    .enum {
        BLACK = 0,
        WHITE = 1,
        RED = 2,
        CYAN = 3,
        PURPLE = 4,
        GREEN = 5,
        BLUE = 6,
        YELLOW = 7
    }
}

.namespace Registers {
    .enum {
        BORDER = $d020,
        BACKGROUND = $d021
    }
}

* = $0801
start:
    lda #Colors.RED
    sta Registers.BORDER
    lda #Colors.BLUE
    sta Registers.BACKGROUND
    rts
