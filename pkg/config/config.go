package config

type CLI struct {
	TLSCertPath string
	TLSKeyPath  string
	DBPath      string
	Insecure    bool
	HttpAddr    string
	HttpsAddr   string
}
