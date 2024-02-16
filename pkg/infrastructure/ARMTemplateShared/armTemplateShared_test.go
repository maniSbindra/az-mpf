package ARMTemplateShared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParametersInStandardFormatWithSchema(t *testing.T) {
	parameters := map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]interface{}{
			"adminUsername": map[string]interface{}{
				"value": "GEN-UNIQUE",
			},
			"adminPasswordOrKey": map[string]interface{}{
				"value": "GEN-PASSWORD",
			},
			"dnsLabelPrefix": map[string]interface{}{
				"value": "GEN-UNIQUE",
			},
		},
	}

	expected := map[string]interface{}{
		"adminUsername": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
		"adminPasswordOrKey": map[string]interface{}{
			"value": "GEN-PASSWORD",
		},
		"dnsLabelPrefix": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
	}

	result := GetParametersInStandardFormat(parameters)
	assert.Equal(t, expected, result)
}

func TestGetParametersInStandardFormatWithoutSchema(t *testing.T) {
	parameters := map[string]interface{}{
		"adminUsername": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
		"adminPasswordOrKey": map[string]interface{}{
			"value": "GEN-PASSWORD",
		},
		"dnsLabelPrefix": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
	}

	expected := map[string]interface{}{
		"adminUsername": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
		"adminPasswordOrKey": map[string]interface{}{
			"value": "GEN-PASSWORD",
		},
		"dnsLabelPrefix": map[string]interface{}{
			"value": "GEN-UNIQUE",
		},
	}

	result := GetParametersInStandardFormat(parameters)
	assert.Equal(t, expected, result)
}
