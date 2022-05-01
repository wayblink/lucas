package cmd

import (
	"github.com/ringtail/lucas/backend"
	"github.com/ringtail/lucas/backend/types"
	log "github.com/sirupsen/logrus"
)

type CommandLine struct {
	Opts *types.Opts
}

func (cl *CommandLine) Run() {
	if &cl.Opts != nil && cl.Opts.DebugMode == true {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	lbe := &backend.LucasServer{}
	lbe.Start(cl.Opts)
}
