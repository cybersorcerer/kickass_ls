// Test Case 07: .enum Directive
// Purpose: Test enum member parsing with values
// Issue #3: Directive Parameter Parsing
//
// .enum takes no name (manual 4.3): .enum {a, b, c}. Members are defined as
// constants in the surrounding scope, so each group is wrapped in a namespace
// to keep the qualified access shown below (manual 9.4).

* = $0801

// Basic enum with explicit values
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

// Enum with auto-increment values
.namespace Sprites {
    .enum {
        PLAYER,      // 0
        ENEMY1,      // 1
        ENEMY2,      // 2
        BULLET       // 3
    }
}

// Enum with mixed values
.namespace States {
    .enum {
        IDLE = 0,
        RUNNING = 1,
        JUMPING = 2,
        FALLING = 3,
        DEAD = 99
    }
}

// Enum with expression values
.namespace Addresses {
    .enum {
        SCREEN = $0400,
        CHARSET = $2000,
        SPRITES = $3000
    }
}

start:
    // Use enum values
    lda #Colors.RED
    sta $d020

    lda #Sprites.PLAYER
    sta $0340

    lda #States.IDLE
    sta $80
    rts
