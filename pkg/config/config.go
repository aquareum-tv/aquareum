package config

type BuildFlags struct {
	Version   string
	BuildTime int64
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
