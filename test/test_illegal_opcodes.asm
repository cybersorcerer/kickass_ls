// Test file for illegal opcodes
.const ZERO_PAGE = $80

start:
    // Test common illegal opcodes
    slo $d020       // Shift Left then OR
    rla ZERO_PAGE   // Rotate Left then AND
    sre #$FF        // Shift Right then EOR
    rra $0400,x     // Rotate Right then Add
    sax $81         // Store A AND X
    lax $82         // Load A and X
    dcp $d021       // Decrease then Compare
    isc $0400,y     // Increment then Subtract with Carry

    // Test additional illegal opcodes
    anc #$80        // AND with Carry flag
    asr #$7F        // AND then Shift Right
    arr #$C0        // AND then Rotate Right
    sbx #$10        // Subtract from X

    // Test NOP variants
    dop $80         // Double NOP
    top $C000       // Triple NOP

    // Test dangerous opcodes
    jam             // Halt CPU

    rts