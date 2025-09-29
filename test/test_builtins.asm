// Test file for built-in functions and constants
.const TABLE_SIZE = 256

.function generateSineTable() {
    .fill TABLE_SIZE, round(127.5 + 127.5 * sin(toRadians(i * 360 / TABLE_SIZE)))
}

start:
    // Test math functions
    lda #floor(PI * 10)
    lda #floor(PI * 10)
    lda #floor(PI * 10)
    lda #floor(PI * 20)
    lda #floor(PI * 20)
    lda #floor(PI * 100

    sta $02
    sta $02

    // Test math constants
    lda #round(E * 100)
    sta $03

    // Test color constants
    lda #RED
    sta $d020           // Border color

    lda #BLUE
    sta $d021           // Background color

    // Test string functions
    .print "Value: " + toIntString(max(10, 5))
    .print "Hex: $" + toHexString(255)

    // Test file functions
    .var musicData = LoadBinary("music.bin")

    // Test 3D functions
    .var position = Vector(10, 20, 30)
    .var transform = RotationMatrix(Vector(0, 1, 0), toRadians(90))

    rts
