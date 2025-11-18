package main

import (
	"encoding/json"
	"fmt"

	invpkg "github.com/pobyzaarif/belajarGo2/app/jsonvsproto/inventory"
	"google.golang.org/protobuf/proto"
)

// ProtoMarshal marshals a protobuf message to binary using the generated proto code.
func ProtoMarshal(inv *invpkg.InventoryRequest) ([]byte, error) {
	return proto.Marshal(inv)
}

// ProtoUnmarshal unmarshals binary protobuf data into the message.
func ProtoUnmarshal(data []byte) (*invpkg.InventoryRequest, error) {
	var inv invpkg.InventoryRequest
	if err := proto.Unmarshal(data, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// JSONMarshal marshals the same struct using encoding/json.
func JSONMarshal(inv *invpkg.InventoryRequest) ([]byte, error) {
	return json.Marshal(inv)
}

// JSONUnmarshal unmarshals JSON data into the struct.
func JSONUnmarshal(data []byte) (*invpkg.InventoryRequest, error) {
	var inv invpkg.InventoryRequest
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

func main() {
	fmt.Println("Run `go test ./app/jsonvsproto/... -bench . -benchmem` to run benchmarks comparing protobuf vs JSON")
}
