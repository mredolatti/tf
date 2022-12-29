package config

import (
	conf "github.com/mredolatti/tf/codigo/common/config"
)

type Main struct {
	Debug               bool
	DBEngine            string
	GoogleCredentialsFn string
	Server              conf.Server
	Mongo               conf.Mongo
	Postgres            conf.Postgres
}
