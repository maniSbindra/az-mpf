# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY AN AZURE CONTAINER Instance
# This is an example of how to deploy an Azure Container Instance
# See test/terraform_azure_aci_example_test.go for how to write automated tests for this code.
# ---------------------------------------------------------------------------------------------------------------------

# ------------------------------------------------------------------------------
# CONFIGURE OUR AZURE CONNECTION
# ------------------------------------------------------------------------------

terraform {
#   required_providers {
#     azurerm = {
#       version = "~>2.29.0"
#       source  = "hashicorp/azurerm"
#     }
#   }

}

provider "azurerm" {
  features {}
  skip_provider_registration = "true"
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY A RESOURCE GROUP
# ---------------------------------------------------------------------------------------------------------------------

# add random id to resource group name to avoid conflicts
resource "random_id" "rg" {
  byte_length = 8
}
resource "azurerm_resource_group" "rg" {
  name     = "rg-${random_id.rg.hex}"
  location = var.location
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY AN AZURE CONTAINER INSTANCE
# ---------------------------------------------------------------------------------------------------------------------

resource "azurerm_container_group" "aci" {
  name                = "aci${var.postfix}"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_address_type = "Public"
  dns_name_label  = "aci${var.postfix}"
  os_type         = "Linux" 

  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "0.5"
    memory = "1.5"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }

  tags = {
    Environment = "Development"
  }
}