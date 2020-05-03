package config

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Setup setups up application configuration
func Setup(ctx context.Context) error {
	// TODO: make this configurable
	logToFile := false
	if logToFile {
		// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.JSONFormatter{})

		logfile := "app.log"
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Fatal("could not open logfile")
		}

		err = os.Truncate(logfile, 0)
		if err != nil {
			log.WithError(err).Fatal("could not truncate file")
		}

		log.SetOutput(f)
	}

	loglevel := "warn"
	level, err := log.ParseLevel(loglevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	return nil
}
