package terraform

import (
	"bytes"
	"strings"
	"testing"

	"github.com/manisbindra/az-mpf/pkg/domain"
)

func TestSaveResultAsJSON(t *testing.T) {
	tests := []struct {
		name       string
		mpfResult  domain.MPFResult
		wantErr    bool
		wantOutput string
	}{
		{
			name: "valid MPFResult",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{
					"": []string{
						"Microsoft.Authorization/roleAssignments/delete",
						"Microsoft.Authorization/roleAssignments/read",
						"Microsoft.Authorization/roleAssignments/write",
					},
				},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{"":["Microsoft.Authorization/roleAssignments/delete","Microsoft.Authorization/roleAssignments/read","Microsoft.Authorization/roleAssignments/write"]}}`,
		},
		{
			name: "valid MPFResult with detailed permissions",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{
					"": []string{
						"Microsoft.ContainerService/managedClusters/read",
						"Microsoft.ContainerService/managedClusters/write",
						"Microsoft.Network/virtualNetworks/read",
						"Microsoft.Network/virtualNetworks/subnets/read",
						"Microsoft.Network/virtualNetworks/subnets/write",
						"Microsoft.Network/virtualNetworks/write",
						"Microsoft.Resources/deployments/read",
						"Microsoft.Resources/deployments/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster": []string{
						"Microsoft.ContainerService/managedClusters/read",
						"Microsoft.ContainerService/managedClusters/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet": []string{
						"Microsoft.Network/virtualNetworks/read",
						"Microsoft.Network/virtualNetworks/write",
					},
					"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet": []string{
						"Microsoft.Network/virtualNetworks/subnets/read",
						"Microsoft.Network/virtualNetworks/subnets/write",
					},
				},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{"":["Microsoft.ContainerService/managedClusters/read","Microsoft.ContainerService/managedClusters/write","Microsoft.Network/virtualNetworks/read","Microsoft.Network/virtualNetworks/subnets/read","Microsoft.Network/virtualNetworks/subnets/write","Microsoft.Network/virtualNetworks/write","Microsoft.Resources/deployments/read","Microsoft.Resources/deployments/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.ContainerService/managedClusters/azmpfakstestcluster":["Microsoft.ContainerService/managedClusters/read","Microsoft.ContainerService/managedClusters/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet":["Microsoft.Network/virtualNetworks/read","Microsoft.Network/virtualNetworks/write"],"/subscriptions/SSSSSSSS-SSSS-SSSS-SSSS-SSSSSSSSSSSS/resourceGroups/testdeployrg-1Gb2X44/providers/Microsoft.Network/virtualNetworks/azmpfakstestvnet/subnets/azmpfakstestsubnet":["Microsoft.Network/virtualNetworks/subnets/read","Microsoft.Network/virtualNetworks/subnets/write"]}}`,
		},
		{
			name: "empty MPFResult",
			mpfResult: domain.MPFResult{
				RequiredPermissions: map[string][]string{},
			},
			wantErr:    false,
			wantOutput: `{"RequiredPermissions":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := saveResultAsJSON(&buf, tt.mpfResult)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveResultAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutput := buf.String(); strings.TrimSpace(gotOutput) != strings.TrimSpace(tt.wantOutput) {
				t.Errorf("saveResultAsJSON() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}
