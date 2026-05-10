package utils

import "testing"

func TestToUpperAtN(t *testing.T) {
	tests := []struct {
		name     string
		position int
		input    string
		want     string
	}{
		{
			name:     "first letter lowercase",
			position: 0,
			input:    "hello",
			want:     "Hello",
		},
		{
			name:     "middle letter",
			position: 2,
			input:    "hello",
			want:     "heLlo",
		},
		{
			name:     "already uppercase",
			position: 0,
			input:    "Hello",
			want:     "Hello",
		},
		{
			name:     "out of bounds",
			position: 10,
			input:    "hello",
			want:     "hello",
		},
		{
			name:     "negative position",
			position: -1,
			input:    "hello",
			want:     "hello",
		},
		{
			name:     "unicode character",
			position: 1,
			input:    "café",
			want:     "cAfé",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToUpperAtN(tt.position, tt.input)
			if got != tt.want {
				t.Errorf("ToUpperAtN(%d, %q) = %q, want %q",
					tt.position, tt.input, got, tt.want)
			}
		})
	}
}
