// Test file for comment completion suppression
// This tests that completions are NOT offered inside comments

.const MYCONST = 10
.var myVar = 20

myLabel:
    nop

// Test 1: Inside // comment - should return EMPTY completion list
// test

// Test 2: After semicolon - should return EMPTY completion list
; test

// Test 3: Inside // comment with partial word - should return EMPTY
// ld

// Test 4: Inside ; comment with partial word - should return EMPTY
; ld

// Test 5: Normal completion after lda (not in comment) - should show completions
lda

// Test 6: Completion at start of line (not in comment) - should show completions

