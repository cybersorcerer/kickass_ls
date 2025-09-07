
.const GREETING = "HELLO"
.const MAX_SPRITES = 8
.var   sprite_count = 0

.function Add(a, b) {
  .return a + b
}

.macro Print(text) {
    .for (var i=0; i<text.size(); i++) {
         lda #text.charAt(i)
         jsr $ffd2 
    }
}

Main:
  ldx #0
  lda #MAX_SPRITES
  sta sprite_count

  lda #Add(5, 3) 
  ldx #$00
  +Print(GREETING)
  +Print(GREETING) 
  rts
