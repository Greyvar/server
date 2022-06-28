package main;

import (
	"time"
	log "github.com/sirupsen/logrus"
	greyvar "github.com/greyvar/server/pkg/greyvarserver"
)

func main() {
	log.Println("Greyvar Server 2");

	greyvar.Start();

	time.Sleep(time.Second * 3)
}
