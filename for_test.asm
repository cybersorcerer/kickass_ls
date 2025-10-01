.for (var i = 0; i < 8; i++) {
    .byte i
}
.for (var x = 0; x < 4; x++) {
    .for (var y = 0; y < 3; y++) {
        .byte x, y
    }
}