package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := add(2, 3)
	if result != 5 {
		t.Errorf("add(2, 3) = %d; want 5", result)
	}
}

func TestSubtract(t *testing.T) {
	result := subtract(5, 3)
	if result != 2 {
		t.Errorf("subtract(5, 3) = %d; want 2", result)
	}
}
