package main

import (
	"encoding/json"
	"fmt"

	v0 "aquareum.tv/aquareum/pkg/schema/v0"
)

func main() {
	err := Main()
	if err != nil {
		panic(err)
	}
}

// Exports the generated EIP-712 schema for use elsewhere
func Main() error {
	schema, err := v0.MakeV0Schema()
	if err != nil {
		return err
	}
	eipSchema, err := schema.EIP712()
	if err != nil {
		return err
	}
	out := map[string]any{
		"domain": eipSchema.Domain,
		"types":  eipSchema.Types,
	}
	bs, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bs))
	return nil
}
