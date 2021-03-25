package core

// tool to find the ping time to an array of ip addresses

// normally used to find the closest data center for a cloud provider.  caller
// passes in an array of ips with some details about the ip and is returned the same
// array with ping times for each ip address

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-ping/ping"
)

const maxPingTime = 3 * time.Second // assumes there will be atleast one region closer than this
const numPings = 3                  // how many pings test should do to an IP

type IPPingTime struct {
	ID      string // id to identify what ip is associated with (normally region id)
	Address string // ip address or hostname that we can use for ping test
	Time    int    // ping time for given ip address (in msec)
}

// getPingTime will do a ping test to the given host / ip address
func getPingTime(pingDetails *IPPingTime, ch chan IPPingTime) {

	pinger, err := ping.NewPinger(pingDetails.Address)
	if err == nil {

		// set ping parameters
		pinger.Count = numPings
		pinger.Timeout = time.Millisecond * maxPingTime // milliseconds

		// do the ping test
		err = pinger.Run() // blocks until finished
		if err == nil {
			// get results
			stats := pinger.Statistics() // get send/receive/rtt stats

			// save results
			pingTime := stats.AvgRtt.Milliseconds()
			pingDetails.Time = int(pingTime)
			log.Debug(pingDetails.ID, " ", pingDetails.Time, " seconds")
		}
	}

	if err != nil {
		// log the error
		log.Warn(err)
	}

	// pass results back to caller
	// note: need to pass something back whether worked or not
	// callers waits until gets results from all ping tests
	ch <- *pingDetails
}

// GetPingTimesForArray will ping all ips/hosts and return the ID of the closest
func GetPingTimesForArray(ipAddressList []IPPingTime) (closestRegion string, err error) {

	numIPs := len(ipAddressList)

	// create a channel for ping tests to return results to us
	// don't want any blocking to make size of number of tests
	ch := make(chan IPPingTime, numIPs)

	// set a timeout as we don't want to hang if
	// don't get all ping results
	f := func() {
		close(ch)
	}

	// issue the pings
	for index := range ipAddressList {
		go getPingTime(&ipAddressList[index], ch)
	}

	// start the timeout timer
	// timeout should be longer than timeout on pings
	timeout := time.AfterFunc(maxPingTime+time.Second, f)
	defer timeout.Stop()

	// reading result until we have them all
	numResults := 0
	lowestPingTime := math.MaxInt32
	for result := range ch {

		numResults++

		// keep track of fastest ping time
		// ignore time of 0 (means ping failed)
		if (result.Time > 0) && (result.Time < lowestPingTime) {
			closestRegion = result.ID
			lowestPingTime = result.Time
		}
		// log.Debug(result.name, result.time)

		// do we have all the results?
		// NOTE: this method was unreliable and would sometimes hang
		//       seems like some of the ping tests would not report back
		// if numResults == numRegions {
		// 	close(ch) // will break the loop
		// }
	}

	return closestRegion, nil
}
