# az-mpf utility (Azure Deployment Minimum Permissions Finder)

This utility finds the minimum permissions required for a given Azure deployment. This can help when you need to figure out the details of what permissions a service principal or managed identity will need to deploy a given ARM template or Terrform module. Similarly when assigning a Service Principal / Managed Identity to an Azure Policy Assignment, this utility can help you figure out the minimum permissions required by the Service Principal / Managed Identity to enforce/remediate the policy. It is recommended that the utility is used in a development or test environment to find the minimum permissions required.

## How It works

The overview of how this utility works is as follows:

![Overview](docs/images/overview.png)

Following is the detailed flow of how this utility works:
* The key parameters the utility needs are the **Service Principal details** (Client ID, Secret and Object ID) and details needed for the specific deployment provider: 
  * ARM: ARM template file and parameters file needed
  * Terraform: Terraform module directory and variables file needed
* The utility **removes any existing Role Assignments for provided Service Principal**
* A Custom Role (with no assigned permissions) is created
* The Service Principal (SP) is assigned the new custom role
* For the above steps the utitity uses the **default Azure CLI credentials** which needs to have permissions to create custom role, and role assignments
* For the following sub steps the **Service Principal Credentials** are used. These sub steps are re-tried till the deployment succeeds
  * Depending on the provider (ARM whatif, ARM fullDeployment, or Terraform) a deployment is tried
  * If the Service Principal does not have sufficient permissions an authorization error is returned. If Authorization errors have occured, parse Authorization errors to fetch the missing scopes and permissions. The [authorizationErrorParser Tests](./pkg/domain/authorizationErrorParser_test.go) provide details of the differnt kinds of Authorization errors typically received.
  * Add the missing permissions to custom Role. This sub step uses **default Azure CLI credentials** to update role permissions.
* Once no authorization error is received, the utility prints the permissions assigned to the Service Principal
* Print the Permissions. There are options provided to view the resource wise break up of permissions and also to export the result in JSON format
* Cleanup up the resource created by the utility including the Role Assignments and Custom Role.

## Supported Deployment Providers

* Azure ARM Template: The ARM endpoints of both modes described below return multiple authorization errors at a time and as a result the final output is usually displayed in under 90 seconds, even for complex templates.
  * Whatif mode: This this the default ARM mode. It uses the ARM whatif endpoint to get the authorization errors and find the minimum permissions required for a deployment, without actually create any resources.
  * FullDeployment mode: This mode creates an ARM template deployment request and then parses the authorization errors to find the minimum permissions required for a deployment. The deployment and resource group are deleted as soon as we get past the authorization errors, and as a result most resources are not actually created. One advantage of this mode is that it is quicker than the whatif mode, as the the responses from the whatif endpoing are not synchronous and take some time to get prepared
* Terraform: The Terraform modes finds the minimum permissions required for a deployment, by getting the authorization errors from the Terraform plan/apply, and destroy commands. All resources are cleaned up by the utility. Since terraform call the ARM apis for one resource at a time, the authorization errors are not received in bulk, and as a result it can take quite long to get the final result. The overall time is the time taken to run the terraform plan/apply/destroy commands, plus the overhead of getting and parsing the authorization errors a few times.
* Bicep (coming soon)

## Flags and Environment Variables

The commands can be used with flags or environment variables. The following are the flags and environment variables supported by the utility:

### Global Flags (Common to all providers)

| Flag               | Environment Variable | Required / Optional | Description                                                                                                      |
| ------------------ | -------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| subscriptionID     | MPF_SUBSCRIPTIONID   | Required           |                                                                                                                  |
| tenantID           | MPF_TENANTID         | Required           |                                                                                                                  |
| spClientID         | MPF_SPCLIENTID       | Required           |                                                                                                                  |
| spObjectID         | MPF_SPOBJECTID       | Required           | Note this is the SP Object id and is different from the Client ID                                                |
| spClientSecret     | MPF_SPCLIENTSECRET   | Required           |                                                                                                                  |
| showDetailedOutput | MPF_SHOWDETAILEDOUTPUT | Optional       | If set to true, the output shows details of permissions resource wise as well. This is not needed if --jsonOutput is specified |
| jsonOutput         | MPF_JSONOUTPUT       | Optional           | If set to true, the detailed output is printed in JSON format                                                            |
| verbose            | MPF_VERBOSE          | Optional           | If set to true, verbose output with informational messages is displayed                                          |
| debug              | MPF_DEBUG            | Optional           | If set to true, output with detailed debug messages is displayed. The debug messages may contain sensitive tokens |

### ARM Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| templateFilePath     | MPF_TEMPLATEFILEPATH     | Required           | ARM template file with path                                                                                       |
| parametersFilePath   | MPF_PARAMETERSFILEPATH   | Required           | ARM template parameters file with path                                                                            |
| resourceGroupNamePfx | MPF_RESOURCEGROUPNAMEPFX | Optional           | Prefix for the resource group name. If not provided, default prefix is testdeployrg. For ARM deployments this temporary resource group is created                                |
| deploymentNamePfx    | MPF_DEPLOYMENTNAMEPFX    | Optional           | Prefix for the deployment name. If not provided, default prefix is testDeploy. For ARM deployments this temporary deployment is created                                |
| location             | MPF_LOCATION             | Optional           | Location for the resource group. If not provided, default location is eastus                                      |
| fullDeployment       | MPF_FULLDEPLOYMENT       | Optional           | If set to true, the utility tries to create all resources while finding the minimum permissions required. If not specified the ARM whatif endpoint is used and no resources are actually created |

### Terraform Flags

| Flag                 | Environment Variable     | Required / Optional | Description                                                                                                      |
| -------------------- | ------------------------ | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| tfPath               | MPF_TFPATH               | Required           | Path to the Terraform executable                                                                             |
| workingDir           | MPF_WORKINGDIR           | Required           | Path to the Terraform module directory                                                                             |
| varFilePath          | MPF_VARFILEPATH          | Required           | Path to the Terraform variables file                                                                             |


## Installation

You can download the latest version for your platform from the [releases](https://github.com/maniSbindra/az-mpf/releases/) link.

For example, to download the latest version for Windows:

```shell
# Please change version in the URL to the latest version
curl -L https://github.com/maniSbindra/az-mpf/releases/download/v0.4.0/az-mpf-windows-amd64.exe -o az-mpf.
chmod +x ./az-mpf
```

And for Mac Arm64:
  
```shell
# Please change version in the URL to the latest version
curl -L https://github.com/maniSbindra/az-mpf/releases/download/v0.4.0/az-mpf-darwin-arm64 -o az-mpf
chmod +x ./az-mpf
```

## Usage Details

### ARM whatif mode (Default mode for ARM)

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./az-mpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write

```

### ARM fullDeployment mode

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID

$ ./az-mpf arm --templateFilePath ./samples/templates/aks-private-subnet.json --parametersFilePath ./samples/templates/aks-private-subnet-parameters.json --fullDeployment

------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerService/managedClusters/read
Microsoft.ContainerService/managedClusters/write
Microsoft.Network/virtualNetworks/read
Microsoft.Network/virtualNetworks/subnets/read
Microsoft.Network/virtualNetworks/subnets/write
Microsoft.Network/virtualNetworks/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
------------------------------------------------------------------------------------------------------------------------------------------

```

### Terraform

```shell
export MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
export MPF_TENANTID=YOUR_TENANT_ID
export MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
export MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
export MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
export MPF_TFPATH=TERRAFORM_EXECUTABLE_PATH

$ ./az-mpf terraform --workingDir `pwd`/samples/terraform/aci --varFilePath `pwd`/samples/terraform/aci/dev.vars.tfvars 
------------------------------------------------------------------------------------------------------------------------------------------
Permissions Required:
------------------------------------------------------------------------------------------------------------------------------------------
Microsoft.ContainerInstance/containerGroups/delete
Microsoft.ContainerInstance/containerGroups/read
Microsoft.ContainerInstance/containerGroups/write
Microsoft.Resources/deployments/read
Microsoft.Resources/deployments/write
Microsoft.Resources/subscriptions/resourcegroups/delete
Microsoft.Resources/subscriptions/resourcegroups/read
Microsoft.Resources/subscriptions/resourcegroups/write
------------------------------------------------------------------------------------------------------------------------------------------

```

Is is also possible to additionally view detailed resource level permissions required as shown in the [display options](docs/display-options.MD) documents.

The blog post [Figuring out the Minimum Permissions Required to Deploy an Azure ARM Template](https://medium.com/microsoftazure/figuring-out-the-minimum-permissions-required-to-deploy-an-azure-arm-template-d1c1e74092fa) provides a more contextual usage scenario for az-mpf.


## Display Options

To view details of display options the utility provides please refer to the [display options](docs/display-options.MD) document.

## Building Locally

You can also build locally by cloning this repo and running `make build`.

## Debugging Locally and Contributing

The [CONTRIBUTING.md](CONTRIBUTING.md) file provides details on how to debug locally and contribute to this project.




