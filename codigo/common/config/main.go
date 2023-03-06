package config

type Server struct {
	Host         string
	Port         int
	RootCAFn     string
	CertChainFn  string
	PrivateKeyFn string
}

type Mongo struct {
	Hosts    []string
	User     string
	Password string
	DB       string
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

type Redis struct {
	Host string
	Port int
	DB int
}
