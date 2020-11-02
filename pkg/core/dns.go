package core

import (
	"fmt"

	"github.com/go-ping/ping"
)

// const maxPingTime = 750

// HostInfo has DNS details of a host
type HostInfo struct {
	Name string // hostname
	ID   string // assigned by dns provider
	IP   string // current IP address assigned to this host
}

// DNSProvider is the interface all DNS providers need to follow
type DNSProvider interface {
	Init() bool
	GetHostRecord(host string) *HostInfo
	UpdateHostRecord(*HostInfo) error
	DeleteHostRecord(*HostInfo) error
}

// PingTime contains results of ping test to ip address
type PingTime struct {
	Name      string // name (optional) associated with the ip address
	IPAddress string // address to test
	Result    int64  // ping time in mSec
}

// GetPingTime will do a ping test to the given ip address
func GetPingTime(ipAddress string) (pingTime int64, err error) {

	pinger, err := ping.NewPinger(ipAddress)
	// pinger.Timeout = time.Millisecond * maxPingTime // milliseconds
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	pinger.Count = 3
	err = pinger.Run() // blocks until finished
	if err != nil {
		return 0, err
	}
	stats := pinger.Statistics() // get send/receive/rtt stats

	pingTime = stats.AvgRtt.Milliseconds()

	return pingTime, nil
}
