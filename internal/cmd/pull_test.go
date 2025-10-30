package cmd

import (
	"testing"
)

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a less than b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b less than a",
			a:        20,
			b:        15,
			expected: 15,
		},
		{
			name:     "equal values",
			a:        10,
			b:        10,
			expected: 10,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        -10,
			expected: -10,
		},
		{
			name:     "negative and positive",
			a:        -5,
			b:        10,
			expected: -5,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        10,
			expected: 0,
		},
		{
			name:     "zero and negative",
			a:        0,
			b:        -10,
			expected: -10,
		},
		{
			name:     "both zero",
			a:        0,
			b:        0,
			expected: 0,
		},
		{
			name:     "large numbers",
			a:        1000000,
			b:        999999,
			expected: 999999,
		},
		{
			name:     "very large difference",
			a:        1,
			b:        1000000,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMinSymmetry(t *testing.T) {
	// Test that min(a, b) == min(b, a)
	testCases := []struct {
		a int
		b int
	}{
		{5, 10},
		{-5, 10},
		{0, 5},
		{100, 50},
	}

	for _, tc := range testCases {
		result1 := min(tc.a, tc.b)
		result2 := min(tc.b, tc.a)
		if result1 != result2 {
			t.Errorf("min is not symmetric: min(%d, %d) = %d, but min(%d, %d) = %d",
				tc.a, tc.b, result1, tc.b, tc.a, result2)
		}
	}
}

func TestMinIdempotence(t *testing.T) {
	// Test that min(a, a) == a
	testValues := []int{0, 1, -1, 100, -100, 999, -999}

	for _, val := range testValues {
		result := min(val, val)
		if result != val {
			t.Errorf("min(%d, %d) = %d, expected %d", val, val, result, val)
		}
	}
}

func TestMinTransitivity(t *testing.T) {
	// Test that if a <= b and b <= c, then min(a, c) == a
	testCases := []struct {
		a int
		b int
		c int
	}{
		{1, 2, 3},
		{-10, 0, 10},
		{5, 10, 15},
	}

	for _, tc := range testCases {
		if tc.a <= tc.b && tc.b <= tc.c {
			result := min(tc.a, tc.c)
			if result != tc.a {
				t.Errorf("Transitivity failed: min(%d, %d) = %d, expected %d", tc.a, tc.c, result, tc.a)
			}
		}
	}
}

func TestMinBoundaryValues(t *testing.T) {
	// Test with boundary values
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "max int and 0",
			a:        int(^uint(0) >> 1), // Max int
			b:        0,
			expected: 0,
		},
		{
			name:     "min int and 0",
			a:        -(int(^uint(0) >> 1)) - 1, // Min int
			b:        0,
			expected: -(int(^uint(0) >> 1)) - 1,
		},
		{
			name:     "1 and max int",
			a:        1,
			b:        int(^uint(0) >> 1),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
