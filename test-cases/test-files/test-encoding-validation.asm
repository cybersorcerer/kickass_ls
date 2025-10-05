// Test encoding validation

.encoding "petscii_upper"      // ✅ Valid
.encoding "invalid_encoding"   // ⚠️ Should warn: unknown encoding

* = $0801
start:
    nop
    rts
