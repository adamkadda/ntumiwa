package middleware

import (
	"github.com/adamkadda/ntumiwa-site/shared/config"
)

func Setup(cfg config.Config) {
	loggingSetup(cfg.Logging)
}
