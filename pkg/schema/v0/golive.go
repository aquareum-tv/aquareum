package v0

type Schema struct {
	GoLive GoLive
}

type GoLive struct {
	Streamer string `json:"streamer"`
	Title    string `json:"title"`
}

// func (c *GoLive) Type() string {
// 	return "GoLive"
// }

// func (c *GoLive) SignerAddress() string {
// 	return c.Signer
// }
