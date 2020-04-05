package config

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// Setup setups up application configuration
func Setup(ctx context.Context) error {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logfile := "app.log"
	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.WithError(err).Fatal("could not open logfile")
	}

	err = os.Truncate(logfile, 0)
	if err != nil {
		logrus.WithError(err).Fatal("could not truncate file")
	}

	// write to logfile
	logrus.SetOutput(f)

	// Only log the warning severity or above.
	loglevel := "debug"
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	return nil
}
