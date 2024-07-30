package v0

import "aquareum.tv/aquareum/pkg/schema"

var Name = "Aquareum"
var Version = "0.0.1"

type V0Schema struct {
	GoLive GoLive
}
type GoLive struct {
	Streamer string `json:"streamer"`
	Title    string `json:"title"`
}

func MakeV0Schema() (schema.Schema, error) {
	return schema.MakeSchema(Name, Version, V0Schema{})
}
