package bench

import (
	"encoding/json"

	invpkg "github.com/pobyzaarif/belajarGo2/app/jsonvsproto/inventory"
	"google.golang.org/protobuf/proto"
)

var Sample = &invpkg.InventoryRequest{
	Code:        "INV001",
	Name:        "Inventory 001",
	Stock:       100,
	Description: "This is a sample inventory item used for benchmarks.",
	Status:      invpkg.InventoryStatus_ACTIVE,
}

func ProtoMarshal(inv *invpkg.InventoryRequest) ([]byte, error) {
	return proto.Marshal(inv)
}

func ProtoUnmarshal(data []byte) (*invpkg.InventoryRequest, error) {
	var inv invpkg.InventoryRequest
	if err := proto.Unmarshal(data, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

func JSONMarshal(inv *invpkg.InventoryRequest) ([]byte, error) {
	return json.Marshal(inv)
}

func JSONUnmarshal(data []byte) (*invpkg.InventoryRequest, error) {
	var inv invpkg.InventoryRequest
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}
