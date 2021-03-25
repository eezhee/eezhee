package main

import (
	"github.com/eezhee/eezhee/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	cmd.Execute()

	// provider digitalocean set apikey xxxx
	//                              default_size micro
	//                              default_region
	// provider digitalocean list regions
	// provider digitalocean list sizes
	// will add credentials to .credentials file
	// SEEMS LONGER.  Would be easier to justde `eezhee login digitalocean xxxxx`

}
