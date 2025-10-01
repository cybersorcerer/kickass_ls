.const valid_byte = $FF
.const invalid_byte = $1FF
.const valid_word = $FFFF
.const invalid_word = $10000

// String validation
.text "Valid string"
.text "Invalid\x escape"

// Memory address validation
lda valid_byte     // Should be OK
lda invalid_byte   // Should warn: value too large for byte
sta $D020         // Should be OK
sta $10000        // Should error: address out of 16-bit range

// List operations validation
.byte $01, $02, $FF, $100  // Should warn: $100 too large for byte
.word $1000, $FFFF, $10000 // Should warn: $10000 too large for word

// Range validation for different contexts
.const zero_page_addr = $00FF
.const non_zero_page = $0200

lda zero_page_addr     // Should suggest zero-page optimization
lda non_zero_page      // Normal addressing

// Function parameter type validation
.const test_abs = abs("invalid")    // Should error: abs expects number
.const test_min = min($FF, "text")  // Should error: min expects numbers

start:
    rts