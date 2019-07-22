package main;

import (
	"fmt"
	"time"
	greyvar "github.com/greyvar/server/pkg/greyvarserver"
)

func main() {
	fmt.Println("Greyvar Server 2");

	greyvar.Start();

	time.Sleep(time.Second * 3)
}
