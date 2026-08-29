package lsp

import "testing"

// labelAddress assembles a tiny program, places a label after the statement
// under test and returns the address the analyzer computed for it. The label
// address minus the origin is the size the statement contributed.
func labelAddress(t *testing.T, src, label string) int64 {
	t.Helper()

	_, ctx, _ := ParseDocument("file:///pc.asm", src)
	if ctx == nil {
		t.Fatal("no analysis context")
	}
	sym, ok := ctx.DefinedLabels[label]
	if !ok {
		t.Fatalf("label %q was not defined by:\n%s", label, src)
	}
	return sym.Address
}

// TestInstructionLengths checks the byte size the analyzer assigns to each
// addressing mode. The sizes come from mnemonic.json, which lists a length per
// mode; they used to be a hardcoded list that reported every branch as three
// bytes and had no entry for TAX at all.
func TestInstructionLengths(t *testing.T) {
	const origin = 0x1000

	tests := []struct {
		name  string
		instr string
		want  int64
	}{
		// implied and accumulator
		{"implied tax", "        tax", 1},
		{"implied rts", "        rts", 1},
		{"implied nop", "        nop", 1},
		{"implied inx", "        inx", 1},
		{"accumulator asl", "        asl", 1},

		// immediate
		{"immediate", "        lda #$01", 2},

		// zero page and absolute
		{"zeropage", "        lda $80", 2},
		{"absolute", "        lda $1234", 3},
		{"zeropage indexed", "        lda $80,x", 2},
		{"absolute indexed", "        lda $1234,x", 3},

		// indirect forms
		{"indexed indirect", "        lda ($80,x)", 2},
		{"indirect indexed", "        lda ($80),y", 2},
		{"jmp indirect", "        jmp ($1234)", 3},
		{"jmp absolute", "        jmp $1234", 3},

		// relative
		{"branch", "        bne skip", 2},

		// data directives
		{"single byte", "        .byte 1", 1},
		{"byte list", "        .byte 1, 2, 3, 4", 4},
		{"single word", "        .word 1", 2},
		{"word list", "        .word 1, 2", 4},
		{"text", "        .text \"abc\"", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := "*=$1000\n" + tc.instr + "\nskip:\n        rts\n"
			got := labelAddress(t, src, "skip") - origin
			if got != tc.want {
				t.Errorf("%q occupies %d bytes, want %d", tc.instr, got, tc.want)
			}
		})
	}
}

// TestBranchDistanceUsesRealLengths is the practical consequence: with branches
// counted as three bytes the distance was overstated and long but valid loops
// were reported as out of range.
func TestBranchDistanceUsesRealLengths(t *testing.T) {
	// 60 two byte instructions between the label and the branch is 120 bytes,
	// comfortably inside the -128 limit. Counting each as three bytes would
	// push it to 180 and produce a false error.
	body := ""
	for i := 0; i < 60; i++ {
		body += "        lda #$00\n"
	}
	src := "*=$1000\nloop:\n" + body + "        bne loop\n"

	_, diags := parseSource(t, src)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Errorf("valid backward branch was rejected: %s", d.Message)
		}
	}
}

// TestForLoopBodyCountedOnce guards the double counting that came from scanning
// the whole document for .for directives on every recursion level: the body was
// added by the normal walk and again by the scan, once per nesting depth.
func TestForLoopBodyCountedOnce(t *testing.T) {
	// A loop with three iterations over a single one byte instruction occupies
	// three bytes.
	src := "*=$1000\n.for (var i = 0; i < 3; i++) {\n        nop\n}\nafter:\n        rts\n"
	got := labelAddress(t, src, "after") - 0x1000
	if got != 3 {
		t.Errorf("three iterations of a one byte body occupy %d bytes, want 3", got)
	}
}
