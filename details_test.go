package terrors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walteh/terrors"
)

func TestFormatJsonForDetail(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    []byte
		ignored  []string
		priority []string
		want     string
		wantErr  bool
	}{
		{
			name:     "Empty input",
			input:    []byte{},
			ignored:  []string{},
			priority: []string{},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Valid input",
			input:    []byte(`{"error": "something went wrong", "code": 500}`),
			ignored:  []string{},
			priority: []string{},
			want:     "code  = 500\nerror = something went wrong",
			wantErr:  false,
		},

		{
			name:     "Ignored fields",
			input:    []byte(`{"error": "something went wrong", "code": 500, "ignored": "ignore me"}`),
			ignored:  []string{"ignored"},
			priority: []string{},
			want:     "code  = 500\nerror = something went wrong",
			wantErr:  false,
		},
		{
			name:     "Priority fields",
			input:    []byte(`{"error": "something went wrong", "code": 500, "priority": "important"}`),
			ignored:  []string{},
			priority: []string{"priority"},
			want:     "priority = important\ncode     = 500\nerror    = something went wrong",
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			input:    []byte(`{"error": "something went wrong", "code": 500`),
			ignored:  []string{},
			priority: []string{},
			want:     "",
			wantErr:  true,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := terrors.FormatJsonForDetail(tt.input, tt.ignored, tt.priority)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
