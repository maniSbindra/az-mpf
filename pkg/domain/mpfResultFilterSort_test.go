package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUniqueSlice(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		slice    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			slice:    []string{},
			expected: []string{},
		},
		{
			name:     "Slice with duplicate values",
			slice:    []string{"apple", "banana", "apple", "orange", "banana"},
			expected: []string{"apple", "banana", "orange"},
		},
		{
			name:     "Slice with unique values",
			slice:    []string{"apple", "banana", "orange"},
			expected: []string{"apple", "banana", "orange"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUniqueSlice(tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMapWithUniqueValues(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		input    map[string][]string
		expected map[string][]string
	}{
		{
			name:     "Empty map",
			input:    map[string][]string{},
			expected: map[string][]string{},
		},
		{
			name: "Map with empty slices",
			input: map[string][]string{
				"key1": {},
				"key2": {},
			},
			expected: map[string][]string{
				"key1": {},
				"key2": {},
			},
		},
		{
			name: "Map with duplicate values",
			input: map[string][]string{
				"key1": {"apple", "banana", "apple", "orange", "banana"},
				"key2": {"apple", "banana", "apple", "orange", "banana"},
			},
			expected: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
		},
		{
			name: "Map with unique values",
			input: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
			expected: map[string][]string{
				"key1": {"apple", "banana", "orange"},
				"key2": {"apple", "banana", "orange"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapWithUniqueValues(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

}
