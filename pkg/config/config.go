package config

import "time"

type BuildFlags struct {
	Version   string
	BuildTime int64
	UUID      string
}

func (b BuildFlags) BuildTimeStr() string {
	ts := time.Unix(b.BuildTime, 0)
	return ts.UTC().Format(time.RFC3339)
}

func (b BuildFlags) BuildTimeStrMillis() string {
	ts := time.Unix(b.BuildTime, 0)
	return ts.UTC().Format("2006-01-02T15:04:05.000Z")
}

type CLI struct {
	TLSCertPath string
	TLSKeyPath  string
	DBPath      string
	Insecure    bool
	HttpAddr    string
	HttpsAddr   string
	AdminSecret string
	Build       *BuildFlags
}
