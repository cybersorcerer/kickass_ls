#!/bin/bash

# Simple script to test completion after jmp

echo "Testing context-aware completion..."
echo ""

# Test 1: After dot - should show only directives
echo "Test 1: After '.' at line 11, char 5 (should show only directives):"
echo "open test-cases/completion-test.asm" | ./kickass_cl/kickass_cl --server ./kickass_ls --root . --interactive 2>/dev/null | grep -A 50 ">"

echo ""
echo "Test 2: After 'jmp ' at line 24, char 8 (should show only labels):"
# This will be tested manually in interactive mode

echo ""
echo "==================================================="
echo "To test manually, run:"
echo "  ./kickass_cl/kickass_cl --server ./kickass_ls --root . --interactive"
echo ""
echo "Then type:"
echo "  open test-cases/completion-test.asm"
echo "  completion 11 5   # After dot - should show ~27 directives"
echo "  completion 24 8   # After jmp - should show ONLY labels"
echo "  quit"
