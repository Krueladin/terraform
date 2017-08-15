output "vnet_id" {
  value = "${azurerm_virtual_network.vnet.id}"
}

output "vnet_name" {
  value = "${azurerm_virtual_network.vnet.name}"
}

output "vnet_location" {
  value = "${azurerm_virtual_network.vnet.location}"
}

output "vnet_address_space" {
  value = "${azurerm_virtual_network.vnet.location}"
}

output "vnet_dns_servers" {
  value = "${azurerm_virtual_network.vnet.dns_servers}"
}

output "vnet_subnet" {
  value = "${azurerm_virtual_network.vnet.subnet}"
}

