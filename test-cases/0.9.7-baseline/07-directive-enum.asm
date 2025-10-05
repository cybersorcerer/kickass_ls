// Test Case 07: .enum Directive
// Purpose: Test enum member parsing with values
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic enum with explicit values
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

// Enum with auto-increment values
.enum Sprites {
    PLAYER,      // 0
    ENEMY1,      // 1
    ENEMY2,      // 2
    BULLET       // 3
}

// Enum with mixed values
.enum States {
    IDLE = 0,
    RUNNING = 1,
    JUMPING = 2,
    FALLING = 3,
    DEAD = 99
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

// Enum with duplicate values (should warn)
.enum BadEnum {
    FIRST = 0,
    SECOND = 0,    // Duplicate value!
    THIRD = 1
}

// Enum with expression values
.enum Addresses {
    SCREEN = $0400,
    CHARSET = $2000,
    SPRITES = $3000,
    DATA = SPRITES + $0800
}
