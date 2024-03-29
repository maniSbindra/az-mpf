package ARMTemplateShared

import "errors"

var ErrInvalidTemplate = errors.New("InvalidTemplate")

type ArmTemplateAdditionalConfig struct {
	TemplateFilePath   string
	ParametersFilePath string
	DeploymentName     string
}

// Get parameters in standard format that is without the schema, contentVersion and parameters fields
func GetParametersInStandardFormat(parameters map[string]interface{}) map[string]interface{} {
	// convert from
	// {
	// 	"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
	// 	"contentVersion": "1.0.0.0",
	// 	"parameters": {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	//   }

	// convert to
	// {
	// 	  "adminUsername": {
	// 		"value": "GEN-UNIQUE"
	// 	  },
	// 	  "adminPasswordOrKey": {
	// 		"value": "GEN-PASSWORD"
	// 	  },
	// 	  "dnsLabelPrefix": {
	// 		"value": "GEN-UNIQUE"
	// 	  }
	// 	}
	if parameters["$schema"] != nil {

		return parameters["parameters"].(map[string]interface{})

	}
	return parameters
}
