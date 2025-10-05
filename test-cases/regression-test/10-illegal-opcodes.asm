// Regression Test 10: Illegal Opcodes Recognition
// Purpose: Ensure illegal opcodes are recognized (with warnings)
// Status: Should PASS with multiple warnings (expected)

* = $0801

start:
    // LAX - Load A and X
    lax #$00
    lax $80
    lax $1000

    // SAX - Store A AND X
    sax $80
    sax $1000

    // DCP - Decrement and Compare
    dcp $80
    dcp $1000

    // ISC - Increment and Subtract with Carry
    isc $80
    isc $1000

    // SLO - Shift Left and OR
    slo $80
    slo $1000

    // RLA - Rotate Left and AND
    rla $80
    rla $1000

    // SRE - Shift Right and EOR
    sre $80
    sre $1000

    // RRA - Rotate Right and Add with Carry
    rra $80
    rra $1000

    // All should generate warnings but parse correctly
    rts

// Note: These are undocumented/illegal 6510 opcodes
// The server should warn about their use but still recognize them
