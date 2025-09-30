; Test file for lexer-parser integration issues
.const SCREEN = $0400

start:
    ; Test inc mnemonic recognition
    inc $d020

    ; Test anonymous labels
+:  nop
    bcc +

    rts