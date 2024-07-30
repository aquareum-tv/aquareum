package mistclient

type MistStreamInfoTrack struct {
	Codec   string `json:"codec,omitempty"`
	Firstms int64  `json:"firstms,omitempty"`
	Idx     int    `json:"idx,omitempty"`
	Init    string `json:"init,omitempty"`
	Lastms  int64  `json:"lastms,omitempty"`
	Maxbps  int    `json:"maxbps,omitempty"`
	Trackid int    `json:"trackid,omitempty"`
	Type    string `json:"type,omitempty"`
	Bps     int    `json:"bps,omitempty"`

	// Audio Only Fields
	Channels int `json:"channels,omitempty"`
	Rate     int `json:"rate,omitempty"`
	Size     int `json:"size,omitempty"`

	// Video Only Fields
	Bframes int `json:"bframes,omitempty"`
	Fpks    int `json:"fpks,omitempty"`
	Height  int `json:"height,omitempty"`
	Width   int `json:"width,omitempty"`
}

type MistPush struct {
	ID           int64
	Stream       string
	OriginalURL  string
	EffectiveURL string
	Stats        *MistPushStats
}

type MistPushStats struct {
	ActiveSeconds int64 `json:"active_seconds"`
	Bytes         int64 `json:"bytes"`
	MediaTime     int64 `json:"mediatime"`
	Tracks        []int `json:"tracks"`
}
