package lsp

import (
	"strings"
	"testing"
)

// evalConst parses ".const RESULT = <expr>" and returns the value the analyzer
// computed for it. The second result reports whether the expression could be
// evaluated at all; the analyzer only records a value in that case.
func evalConst(t *testing.T, expr string) (int64, bool) {
	t.Helper()

	src := ".const RESULT = " + expr + "\n"
	_, ctx, diags := ParseDocument("file:///expr.asm", src)

	if errs := errorsOnly(diags); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("%q produced an error: %s", expr, e.Message)
		}
		return 0, false
	}
	if ctx == nil {
		t.Fatalf("%q: no analysis context", expr)
	}

	sym, ok := ctx.DefinedLabels["RESULT"]
	if !ok {
		t.Fatalf("%q: constant RESULT was not defined", expr)
	}
	return sym.Address, sym.Value != ""
}

// --- operator coverage ---------------------------------------------------

// TestExpressionOperatorsParse walks every operator listed in the manual
// (tables 4.4, 5.1 and 5.2) plus the two grouping forms. None of them may
// produce a parse error.
func TestExpressionOperatorsParse(t *testing.T) {
	exprs := []string{
		// table 4.4, arithmetic and bitwise
		"10+2", "10-8", "2*3", "10/2",
		">$1020", "<$1020",
		"2<<2", "2>>1",
		"$3f & $0f", "$0f | $30", "$ff ^ $f0", "~%11",
		// table 5.1, comparison
		"1==2", "1!=2", "1>2", "1<2", "1>=2", "1<=2",
		// table 5.2, boolean
		"!1", "1&&2", "1||2",
		// grouping, manual 4.5
		"(2+5)*2", "[2+5]*2", "(([((2+4))])*3)+25",
		// conditional operator, manual 5.1
		"[20<10] ? 1 : 2",
	}

	for _, expr := range exprs {
		t.Run(expr, func(t *testing.T) {
			src := ".const RESULT = " + expr + "\n"
			assertNoErrors(t, src)
		})
	}
}

// --- precedence ----------------------------------------------------------

// TestExpressionPrecedence pins the binding order for the operators the
// analyzer can fold. The expected values follow the Java ordering the manual
// refers to as "standard precedence rules".
func TestExpressionPrecedence(t *testing.T) {
	tests := []struct {
		expr string
		want int64
	}{
		// manual 4.5, verbatim
		{"2+5*2", 12},
		{"(2+5)*2", 14},
		{"[2+5]*2", 14},

		// manual 4.4, verbatim
		{"10+2", 12},
		{"10-8", 2},
		{"2*3", 6},
		{"10/2", 5},
		{"2<<2", 8},
		{"2>>1", 1},
		{"$3f & $0f", 0xf},
		{"$0f | $30", 0x3f},
		{"$ff ^ $f0", 0x0f},

		// left associativity
		{"8/4/2", 1},
		{"10-3-2", 5},

		// * / bind tighter than + -
		{"1+2*3", 7},
		{"2*3+1", 7},

		// + - bind tighter than the shifts
		{"1<<2+1", 8},

		// & tighter than ^, ^ tighter than |
		{"1 | 2 ^ 3 & 4", 3},

		// nesting from the manual
		{"(([((2+4))])*3)+25", 43},
	}

	for _, tc := range tests {
		t.Run(tc.expr, func(t *testing.T) {
			got, evaluated := evalConst(t, tc.expr)
			if !evaluated {
				t.Fatalf("%q could not be evaluated", tc.expr)
			}
			if got != tc.want {
				t.Errorf("%q = %d, want %d", tc.expr, got, tc.want)
			}
		})
	}
}

// --- addressing modes ----------------------------------------------------

// TestJumpParenthesesAreSignificant covers manual 4.5: soft parentheses select
// indirect addressing, hard parentheses only group. The page boundary warning
// must therefore fire for jmp ($10ff) and stay silent for the two absolute
// forms, which the parser used to be unable to tell apart.
func TestJumpParenthesesAreSignificant(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantWarn bool
	}{
		{"indirect through a page boundary", "        jmp ($10ff)\n", true},
		{"absolute", "        jmp $10ff\n", false},
		{"absolute with hard parentheses", "        jmp [$10ff]\n", false},
		{"indirect away from a page boundary", "        jmp ($1000)\n", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := parseSource(t, tc.src)
			got := false
			for _, d := range diags {
				if strings.Contains(d.Message, "page-boundary bug") {
					got = true
				}
			}
			if got != tc.wantWarn {
				t.Errorf("page boundary warning = %v, want %v for %q\n  diagnostics: %v",
					got, tc.wantWarn, tc.src, messages(diags))
			}
		})
	}
}

// --- preprocessor conditions ---------------------------------------------

// TestPreprocessorConditionExpressions covers manual 8.5: #if, #elif and
// #importif take a boolean expression, not just a bare symbol.
func TestPreprocessorConditionExpressions(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"bare symbol", "#define DEBUG\n#if DEBUG\n    nop\n#endif\n"},
		{"negated symbol", "#define DEBUG\n#if !DEBUG\n    nop\n#endif\n"},
		{"and", "#define A\n#define B\n#if A && B\n    nop\n#endif\n"},
		{"or with parentheses", "#define X\n#define Y\n#if X || (X && Y)\n    nop\n#endif\n"},
		{"manual example", "#define DEBUG\n#define COMPLICATED\n#if !DEBUG && !COMPLICATED\n    nop\n#endif\n"},
		{"comparison", "#define DEBUG\n#define X\n#if X==DEBUG\n    nop\n#endif\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertNoErrors(t, tc.src)
		})
	}
}

// --- negative cases ------------------------------------------------------

// TestExpressionErrors checks that broken expressions are reported rather than
// silently swallowed.
func TestExpressionErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"missing closing parenthesis", ".const A = (1 + 2\n"},
		{"missing closing bracket", ".const A = [1 + 2\n"},
		{"conditional without colon", ".const A = 1 ? 2\n"},
		{"operator without right operand", ".const A = 1 +\n"},
		{"invalid hex literal", ".const A = $0G\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := parseSource(t, tc.src)
			if len(errorsOnly(diags)) == 0 {
				t.Errorf("broken expression was accepted: %q\n  diagnostics: %v", tc.src, messages(diags))
			}
		})
	}
}
