package config

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config is application configuration settings
type Config struct {
	LogToFile         bool   `yaml:"log_to_file"`
	LogFile           string `yaml:"log_file"`
	LogLevel          string `yaml:"log_level"`
	Columns           []string
	KeyboardShortcuts map[string]string

	loggers []io.Writer
}

// DefaultConfig is (you guessed it) default application config.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  "warn",
		LogToFile: false,
		LogFile:   "grump.log",
		Columns: []string{
			"artist",
			"album",
			"title",
			"rating",
		},
	}
}

// Setup setups up application configuration
func Setup(ctx context.Context) (*Config, error) {
	c, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: make this configurable
	if c.LogToFile && c.LogFile != "" {
		// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.JSONFormatter{})

		logfile := c.LogFile
		// TODO: close this
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Fatal("could not open logfile")
		}

		err = os.Truncate(logfile, 0)
		if err != nil {
			log.WithError(err).Fatal("could not truncate file")
		}

		// keep this file logger around
		c.loggers = append(c.loggers, f)

		log.SetOutput(f)
	}

	if c.LogFile == "" {
		return c, nil
	}

	level, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		log.WithError(err).Warn("could not set log level")
	} else {
		log.SetLevel(level)
	}

	return c, nil
}

// loadConfig looks for configuration and loads it
func loadConfig(ctx context.Context) (*Config, error) {
	// look for logfile
	var c *Config

	c, err := homeConfig(ctx)
	if err != nil {
		// use default config if something went wrong
		c = DefaultConfig()
	}

	c.loggers = []io.Writer{}

	return c, nil
}

func homeConfig(ctx context.Context) (*Config, error) {
	c := &Config{}

	usr, err := user.Current()
	if err != nil {
		log.WithError(err).Warn("could not obtain current user")
		return nil, err
	}

	homeConfig := filepath.Join(usr.HomeDir, ".grump.yaml")
	b, err := ioutil.ReadFile(homeConfig)
	if err != nil {
		log.WithError(err).WithField("path", homeConfig).Debug("could not read config file")
		return nil, err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		log.WithError(err).WithField("path", homeConfig).Warn("could not unmarshal config yaml")
		return nil, err
	}

	return c, nil
}

// Loggers returns a slice of loggers to use in the application
func (c *Config) Loggers() []io.Writer {
	return c.loggers
}
