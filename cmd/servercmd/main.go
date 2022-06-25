package main;

import (
	"time"
	greyvar "github.com/greyvar/server/pkg/greyvarserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel);
	log.Info("Greyvar Server 2");

	greyvar.Start();

	time.Sleep(time.Second * 3)
}
