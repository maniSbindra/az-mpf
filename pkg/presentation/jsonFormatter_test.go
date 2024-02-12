package presentation

// func TestDisplayJSON(t *testing.T) {
// 	d := &displayConfig{
// 		result: domain.MPFResult{
// 			RequiredPermissions: map[string][]string{
// 				"/subscriptions/86add727-3be8-4151-9c57-cb03117b3f30/resourceGroups/testdeployrg-5VUSYli": {
// 					"Microsoft.Authorization/roleAssignments/read", "Microsoft.Authorization/roleAssignments/write",
// 					"Microsoft.Compute/virtualMachines/extensions/read",
// 					"Microsoft.Compute/virtualMachines/extensions/write",
// 					"Microsoft.Compute/virtualMachines/read",
// 					"Microsoft.Compute/virtualMachines/write",
// 					"Microsoft.ContainerRegistry/registries/read",
// 					"Microsoft.ContainerRegistry/registries/write",
// 					"Microsoft.ContainerService/managedClusters/read",
// 					"Microsoft.ContainerService/managedClusters/write"},
// 				"/subscriptions/86add727-3be8-4151-9c57-cb03117b3f30/resourceGroups/testdeployrg-5VUSYli/providers/Microsoft.Authorization/roleAssignments/c5f7c2f3-ccff-5280-91fb-ea700259559b": {"Microsoft.Authorization/roleAssignments/read", "Microsoft.Authorization/roleAssignments/write"},
// 				"write": {"write", "read"},
// 			},
// 		},
// 	}

// 	var w bytes.Buffer
// 	err := d.displayJSON(&w)
// 	assert.Nil(t, err)

// 	var result map[string][]string
// 	err = json.Unmarshal(w.Bytes(), &result)
// 	assert.Nil(t, err)

// }
