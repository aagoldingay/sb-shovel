terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.68.0"
    }
  }
}

provider "azurerm" {
  features {}
}
