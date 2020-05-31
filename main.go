package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dhulihan/grump/internal/config"
	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/dhulihan/grump/ui"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	ctx := context.Background()

	c, err := config.Setup(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("could not set up config")
	}

	if len(os.Args) < 2 {
		help()
	}

	path := os.Args[1]
	logrus.WithField("path", path).Info("starting up")

	audioShelf, err := library.NewLocalAudioShelf(path)
	if err != nil {
		logrus.WithError(err).Fatal("could not set up audio library")
	}

	count, err := audioShelf.LoadTracks()
	if err != nil {
		logrus.WithError(err).Fatal("could not load audio library")
	}
	logrus.WithField("count", count).Info("loaded library")

	audioShelves := []library.AudioShelf{audioShelf}
	db, err := library.NewLibrary(audioShelves)
	if err != nil {
		logrus.WithError(err).Fatal("could not set up player db")
	}

	player, err := player.NewBeepAudioPlayer()
	if err != nil {
		logrus.WithError(err).Fatal("could not set up audio player")
	}

	build := ui.BuildInfo{
		Version: version,
		Commit:  commit,
	}

	err = ui.Start(ctx, build, db, player, c.Loggers())
	if err != nil {
		logrus.WithError(err).Fatal("ui exited with an error")
	}
}

func help() {
	cmd := os.Args[0]
	fmt.Printf("%s <file or directory>\n", cmd)
	os.Exit(2)
}
