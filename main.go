package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dhulihan/ghoulish/internal/config"
	"github.com/dhulihan/ghoulish/library"
	"github.com/dhulihan/ghoulish/player"
	"github.com/dhulihan/ghoulish/ui"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	err := config.Setup(ctx)
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

	count, err := audioShelf.Scan()
	if err != nil {
		logrus.WithError(err).Fatal("could not scan audio library")
	}
	logrus.WithField("count", count).Info("scanned library")

	audioShelves := []library.AudioShelf{audioShelf}
	db, err := library.NewLibrary(audioShelves)
	if err != nil {
		logrus.WithError(err).Fatal("could not set up player db")
	}

	player, err := player.NewBeepAudioPlayer()
	if err != nil {
		logrus.WithError(err).Fatal("could not set up audio player")
	}

	err = ui.Start(ctx, db, player)
	if err != nil {
		logrus.WithError(err).Fatal("ui exited with an error")
	}
}

func help() {
	cmd := os.Args[0]
	fmt.Printf("%s <file or directory>\n", cmd)
	os.Exit(2)
}
