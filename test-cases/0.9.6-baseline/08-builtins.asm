// Test Case 08: Built-in Functions and Constants
// Purpose: Test Kick Assembler built-in functions and constants

* = $0810

// Built-in constants
.var pi_value = PI
.var e_value = E

// Math functions
.var sine = sin(0.5)
.var cosine = cos(0.5)
.var tangent = tan(0.5)
.var arc_sine = asin(0.5)
.var arc_cosine = acos(0.5)
.var arc_tangent = atan(0.5)
.var square_root = sqrt(16)
.var power = pow(2, 8)
.var absolute = abs(-5)
.var floor_val = floor(3.7)
.var ceil_val = ceil(3.2)
.var round_val = round(3.5)
.var min_val = min(5, 10)
.var max_val = max(5, 10)

// Trigonometric conversions
.var to_rad = toRadians(180)
.var to_deg = toDegrees(PI)

// Random number
.var random = random()

// Logarithmic functions
.var log_val = log(10)
.var log10_val = log10(100)

// Modulo operation
.var mod_val = mod(10, 3)

// List functions
.var list_size = List().size()
.var hash_map = Hashtable()

// String functions
.var str = "HELLO"
.var str_len = str.size()

// Pseudo-random sequence
.var seed = random()

// Binary/Hex helpers
.var byte_val = >$1234      // High byte
.var low_byte = <$1234      // Low byte

// Using built-ins in code
start:
    lda #<$1234
    sta $fb
    lda #>$1234
    sta $fc

// Shift operations
.var shifted = $0400 >> 8

// Conditional expression
.var conditional = (5 > 3) ? $ff : $00

    rts
