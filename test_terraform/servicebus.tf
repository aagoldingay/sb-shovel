resource "azurerm_resource_group" "rg" {
  name     = var.rg_name
  location = var.rg_location

  tags = local.tags
}

resource "azurerm_servicebus_namespace" "sb" {
  name                = var.sb_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = "Basic"

  tags = local.tags
}

resource "azurerm_servicebus_queue" "q" {
  name         = var.q_name
  namespace_id = azurerm_servicebus_namespace.sb.id

  enable_batched_operations = true
  default_message_ttl       = "PT30M" # 30 minutes
}
