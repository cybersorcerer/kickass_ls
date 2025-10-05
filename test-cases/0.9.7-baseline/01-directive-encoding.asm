// Test Case 01: .encoding Directive
// Purpose: Test string parameter parsing for .encoding directive
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic encoding directive with string parameter
.encoding "petscii_upper"

// Test with different encoding types
.encoding "screencode_mixed"
.encoding "petscii_mixed"
.encoding "ascii"

// Program start
start:
    lda #$00
    sta $d020
    rts

// Invalid encoding (should warn)
.encoding "invalid_encoding_name"
