package config

type CLI struct {
	TLSCertPath string
	TLSKeyPath  string
	Insecure    bool
	HttpAddr    string
	HttpsAddr   string
}
