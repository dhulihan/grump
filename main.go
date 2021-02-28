package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	_ "net/http/pprof"

	"github.com/dhulihan/grump/internal/config"
	"github.com/dhulihan/grump/library"
	"github.com/dhulihan/grump/player"
	"github.com/dhulihan/grump/ui"
	log "github.com/sirupsen/logrus"
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
		log.WithError(err).Fatal("could not set up config")
	}

	if len(os.Args) < 2 {
		help()
	}

	path := os.Args[1]
	log.WithField("path", path).Info("starting up")

	// enable profiling
	if c.CPUProfile != "" {
		log.WithField("file", c.CPUProfile).Debug("starting cpu profile")
		f, err := os.Create(c.CPUProfile)
		if err != nil {
			log.WithError(err).Fatal("could not create cpu profile file")
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.WithError(err).Fatal("could not start cpu profile")
		}
		defer pprof.StopCPUProfile()
	}

	audioShelf, err := library.NewLocalAudioShelf(path)
	if err != nil {
		log.WithError(err).Fatal("could not set up audio library")
	}

	count, err := audioShelf.LoadTracks()
	if err != nil {
		log.WithError(err).Fatal("could not load audio library")
	}
	log.WithField("count", count).Info("loaded library")

	audioShelves := []library.AudioShelf{audioShelf}
	db, err := library.NewLibrary(audioShelves)
	if err != nil {
		log.WithError(err).Fatal("could not set up player db")
	}

	player, err := player.NewBeepAudioPlayer()
	if err != nil {
		log.WithError(err).Fatal("could not set up audio player")
	}

	build := ui.BuildInfo{
		Version: version,
		Commit:  commit,
	}

	err = ui.Start(ctx, build, db, player, c.Loggers())
	if err != nil {
		log.WithError(err).Fatal("ui exited with an error")
	}

	// wrap up mem profile
	if c.MemProfile != "" {
		log.WithField("file", c.MemProfile).Debug("finishing mem profile")
		f, err := os.Create(c.MemProfile)
		if err != nil {
			log.WithError(err).Fatal("could not create mem profile file")
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.WithError(err).Fatal("could not write mem profile file")
		}
	}
}

func help() {
	cmd := os.Args[0]
	fmt.Printf("%s <file or directory>\n", cmd)
	os.Exit(2)
}
