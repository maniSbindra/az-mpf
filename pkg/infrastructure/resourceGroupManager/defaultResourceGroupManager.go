package resourcegroupmanager

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureAPI "github.com/manisbindra/az-mpf/pkg/infrastructure/azureAPI"
)

type RGManager struct {
	rgAPIClient *armresources.ResourceGroupsClient
}

func NewResourceGroupManager(subscriptionID string) *RGManager {
	azAPIClient := azureAPI.NewAzureAPIClients(subscriptionID)
	return &RGManager{
		rgAPIClient: azAPIClient.ResourceGroupsClient,
	}
}

func (r *RGManager) DeleteResourceGroup(ctx context.Context, rgName string) error {

	_, err := r.rgAPIClient.BeginDelete(ctx, rgName, nil)
	if err != nil {
		return err
	}

	return nil
}

// method to create resource group
func (r *RGManager) CreateResourceGroup(ctx context.Context, rgName string, location string) error {

	rgParams := armresources.ResourceGroup{
		Location: &location,
		Name:     &rgName,
	}

	// create resource group
	_, err := r.rgAPIClient.CreateOrUpdate(ctx, rgName, rgParams, nil)
	if err != nil {
		return err
	}

	return nil
}
