# Known Issues and Workarounds

## Terraform

### Existing Resource / Import Errors

Creation of certain resources, like the application insights resource, also involves the creation of current billing features resource, as described in the [GitHub issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2467392936). This means that if the identity used by Terraform has permissions to create the application insights resource but not the current billing features resource, the application insights resource will be created in Azure, but the Terraform apply will fail. This means that the creation of the application insights resource will not be tracked by Terraform and the state file will be out of sync with the actual resources in Azure. When the utility adds the required permission and executes Terraform apply again, Terraform will give an error that the resource already exists in Azure and that this can be resolved by importing the resource into the Terraform state file. In this scenario, by default, the MPF utility will import such resources into the Terraform state file and proceed further. It needs to be noted that in the cleanup phase the imported resources will also be deleted.

### Billing Features Payload Error

This issue is also related to the the [Github Issue](https://github.com/hashicorp/terraform-provider-azurerm/issues/27961#issuecomment-2520407658). The utility retries the request to workaround this issue.

