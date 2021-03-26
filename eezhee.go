package main

import (
	"github.com/eezhee/eezhee/cmd"
)

func main() {

	cmd.Execute()

	// provider digitalocean set apikey xxxx
	//                              default_size micro
	//                              default_region
	// provider digitalocean list regions
	// provider digitalocean list sizes
	// will add credentials to .credentials file
	// SEEMS LONGER.  Would be easier to justde `eezhee login digitalocean xxxxx`
}
