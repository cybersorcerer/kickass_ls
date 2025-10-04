// Test Case 06: Context-Aware Completion
// Purpose: Test completion in different contexts

.var test_var = $80
.const TEST_CONST = $d020

* = $0810

// Test 1: Completion after directive (should show nothing before =)
.var new_variable =

// Test 2: Completion after directive = (should show values)
.var another_var = $

// Test 3: Completion after mnemonic (should show addressing modes)
    lda

// Test 4: Completion with # (immediate mode)
    lda #

// Test 5: Completion with $ (absolute address)
    sta $

// Test 6: Completion with ( (indirect mode)
    lda (

// Test 7: Directive completion (type . to trigger)


// Test 8: Mnemonic completion (type l to trigger)


// Test 9: Completion with existing operands
    lda #$00
    lda #$ff
    lda #$80
    lda

// Test 10: C64 memory completion
    lda $d0

// Test 11: Function completion
.function test() {
    .return 0
}

// Test 12: If directive (expects expression)
.if
