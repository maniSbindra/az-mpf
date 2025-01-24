package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDeleteActionFromInvalidActionOrNotActionError(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedActions     []string
		expectedError       bool
		expectedErrorString string
	}{
		{
			name:            "Valid delete action",
			input:           `{"error":{"code":"InvalidActionOrNotAction","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions: []string{"Microsoft.Insights/components/currentbillingfeatures/delete"},
			expectedError:   false,
		},
		{
			name:                "Non InvalidActionOrNotAction error",
			input:               `{"error":{"code":"SomeOtherError","message":"'Microsoft.Insights/components/currentbillingfeatures/delete' does not match any of the actions supported by the providers."}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "Could not parse deploment error, potentially due to a Non-InvalidActionOrNotAction error",
		},
		{
			name:                "No matches found",
			input:               `{"error":{"code":"InvalidActionOrNotAction","message":"No delete action here"}}`,
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
		{
			name:                "Empty input",
			input:               "",
			expectedActions:     []string{},
			expectedError:       true,
			expectedErrorString: "No matches found in 'invalidActionErrorMessage' error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions, err := GetDeleteActionFromInvalidActionOrNotActionError(tt.input)
			assert.Equal(t, tt.expectedActions, actions)
			if tt.expectedError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
