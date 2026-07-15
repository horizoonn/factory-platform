package api

import _ "embed"

//go:embed bundles/order.openapi.v1.bundle.yaml
var orderOpenAPIV1 string

//go:embed generated/inventory/v1/inventory.swagger.json
var inventoryOpenAPIV1 string

func OrderOpenAPIV1() string {
	return orderOpenAPIV1
}

func InventoryOpenAPIV1() string {
	return inventoryOpenAPIV1
}
