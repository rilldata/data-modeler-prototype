package formatter

import (
	"fmt"
	"testing"
)

func TestPerRangeFormatterDefaultGenericNums(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		// integers
		{999_999_999, "1.0B"},
		{12_345_789, "12.3M"},
		{2_345_789, "2.3M"},
		{999_999, "1.0M"},
		{345_789, "345.8k"},
		{45_789, "45.8k"},
		{5_789, "5.8k"},
		// note: no trailing dot
		{999, "999"},
		{789, "789"},
		{89, "89"},
		{9, "9"},
		{0, "0"},
		{-0, "0"},
		{-999_999_999, "-1.0B"},
		{-12_345_789, "-12.3M"},
		{-2_345_789, "-2.3M"},
		{-999_999, "-1.0M"},
		{-345_789, "-345.8k"},
		{-45_789, "-45.8k"},
		{-5_789, "-5.8k"},
		{-999, "-999"},
		{-789, "-789"},
		{-89, "-89"},
		{-9, "-9"},

		// non integers
		{999_999_999.1234686, "1.0B"},
		{12_345_789.1234686, "12.3M"},
		{2_345_789.1234686, "2.3M"},
		{999_999.4397, "1.0M"},
		{345_789.1234686, "345.8k"},
		{45_789.1234686, "45.8k"},
		{5_789.1234686, "5.8k"},
		{999.999, "1.0k"},
		{999.995, "1.0k"},
		{999.994, "999.99"},
		{999.99, "999.99"},
		{999.1234686, "999.12"},
		{789.1234686, "789.12"},
		{89.1234686, "89.12"},
		{9.1234686, "9.12"},
		{0.1234686, "0.12"},

		{-999_999_999.1234686, "-1.0B"},
		{-12_345_789.1234686, "-12.3M"},
		{-2_345_789.1234686, "-2.3M"},
		{-999_999.4397, "-1.0M"},
		{-345_789.1234686, "-345.8k"},
		{-45_789.1234686, "-45.8k"},
		{-5_789.1234686, "-5.8k"},
		{-999.999, "-1.0k"},
		{-999.1234686, "-999.12"},
		{-789.1234686, "-789.12"},
		{-89.1234686, "-89.12"},
		{-9.1234686, "-9.12"},
		{-0.1234686, "-0.12"},

		// no padding with insignificant zeros
		{999.9, "999.9"},
		{999.1, "999.1"},
		{789.1, "789.1"},
		{89.9, "89.9"},
		{9.9, "9.9"},

		// infinitesimals , no padding with insignificant zeros
		{0.9, "0.9"},
		// note 0.10 IS significant
		{0.095, "0.10"},
		{0.0095, "0.01"},
		{0.001, "1.0e-3"},
		{0.00095, "950.0e-6"},
		{0.000999999, "1.0e-3"},
		{0.00012335234, "123.4e-6"},
		{0.000_000_999999, "1.0e-6"},
		{0.000_000_02341253, "23.4e-9"},
		{0.000_000_000_999999, "1.0e-9"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			formatter, err := newPerRangeFormatter(defaultGenericNumOptions())
			if err != nil {
				t.Errorf("failed: %v", err)
			}
			if got, _ := formatter.stringFormat(tt.input); got != tt.expected {
				t.Errorf("perRangeFormatter.stringFormat(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
