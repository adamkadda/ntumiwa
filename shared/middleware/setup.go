package middleware

import (
	"github.com/adamkadda/ntumiwa/shared/config"
)

func Setup(cfg config.Config) {
	loggingSetup(cfg.Logging)
}
