package main

import "testing"

func TestValidateFormatInput(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		expectErr bool
	}{
		{"valid short format", "short", false},
		{"valid format at limit", string(make([]byte, 1024)), false},
		{"invalid too long", string(make([]byte, 1025)), true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormatInput(tt.format)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateFormatInput(%q) error = %v, expectErr %v", tt.format, err, tt.expectErr)
			}
		})
	}
}
