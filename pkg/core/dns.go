package core

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
