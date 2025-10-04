// Test Case 04: Illegal/Undocumented Opcodes
// Purpose: Test support for 6510 illegal opcodes

* = $0810

// LAX - Load A and X
    lax #$00        // Immediate
    lax $80         // Zero Page
    lax $80, y      // Zero Page, Y
    lax $1000       // Absolute
    lax $1000, y    // Absolute, Y
    lax ($80, x)    // Indexed Indirect
    lax ($80), y    // Indirect Indexed

// SAX - Store A AND X
    sax $80         // Zero Page
    sax $80, y      // Zero Page, Y
    sax $1000       // Absolute
    sax ($80, x)    // Indexed Indirect

// DCP - Decrement and Compare
    dcp $80         // Zero Page
    dcp $80, x      // Zero Page, X
    dcp $1000       // Absolute
    dcp $1000, x    // Absolute, X
    dcp $1000, y    // Absolute, Y
    dcp ($80, x)    // Indexed Indirect
    dcp ($80), y    // Indirect Indexed

// ISC - Increment and Subtract with Carry
    isc $80         // Zero Page
    isc $80, x      // Zero Page, X
    isc $1000       // Absolute
    isc $1000, x    // Absolute, X
    isc $1000, y    // Absolute, Y
    isc ($80, x)    // Indexed Indirect
    isc ($80), y    // Indirect Indexed

// SLO - Shift Left and OR
    slo $80         // Zero Page
    slo $80, x      // Zero Page, X
    slo $1000       // Absolute
    slo $1000, x    // Absolute, X
    slo $1000, y    // Absolute, Y
    slo ($80, x)    // Indexed Indirect
    slo ($80), y    // Indirect Indexed

// RLA - Rotate Left and AND
    rla $80         // Zero Page
    rla $80, x      // Zero Page, X
    rla $1000       // Absolute
    rla $1000, x    // Absolute, X
    rla $1000, y    // Absolute, Y
    rla ($80, x)    // Indexed Indirect
    rla ($80), y    // Indirect Indexed

// SRE - Shift Right and EOR
    sre $80         // Zero Page
    sre $80, x      // Zero Page, X
    sre $1000       // Absolute
    sre $1000, x    // Absolute, X
    sre $1000, y    // Absolute, Y
    sre ($80, x)    // Indexed Indirect
    sre ($80), y    // Indirect Indexed

// RRA - Rotate Right and Add with Carry
    rra $80         // Zero Page
    rra $80, x      // Zero Page, X
    rra $1000       // Absolute
    rra $1000, x    // Absolute, X
    rra $1000, y    // Absolute, Y
    rra ($80, x)    // Indexed Indirect
    rra ($80), y    // Indirect Indexed
