package daemon

import (
	"time"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func logger() {
	if Config.logFormat == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr,
		    TimeFormat: time.RFC3339})
	}
}
