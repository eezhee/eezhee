package cloudflare

import (
	"context"
	"errors"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/eezhee/eezhee/pkg/core"
	log "github.com/sirupsen/logrus"
)

// DNSManager controls dns records for hosts
type DNSManager struct {
	APIToken string
	api      *cf.API
}

// NewDNSManager create an object to manage dns
func NewDNSManager(providerAPIToken string) (DNSManager, error) {

	var manager DNSManager

	// make sure api token provided
	if len(providerAPIToken) == 0 {
		return manager, errors.New("API token not specified")
	}

	manager.APIToken = providerAPIToken

	worked := manager.Init()
	if !worked {
		return manager, errors.New("could not init")
	}

	// dns manager ready to be used
	return manager, nil
}

// Test allows us to get familiar with the Cloudflare API
func Test(apiToken string) bool {

	// make sure we have an API token, else return
	manager, err := NewDNSManager(apiToken)
	if err != nil {
		return false
	}

	clusterName := "cluster1.k8s.rndguy.ca"
	clusterDNSRecord := core.HostInfo{Name: clusterName, IP: "218.1.2.3"}

	hostInfo, err := manager.GetHostRecord(clusterName)
	if err != nil {
		log.Error(err)
		return false
	}
	log.Debug(hostInfo)

	err = manager.AddHostRecord(clusterDNSRecord)
	if err != nil {
		log.Error(err)
		return false
	}
	log.Debug(hostInfo)

	clusterDNSRecord.IP = "218.1.2.4"
	err = manager.UpdateHostRecord(clusterDNSRecord)
	if err != nil {
		log.Error(err)
		return false
	}

	err = manager.DeleteHostRecord(clusterDNSRecord)
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

// Init will do final config of manager
func (m *DNSManager) Init() bool {

	// init cloudflare package
	// m.api, err := cf.NewWithAPIToken(m.APIToken)
	api, err := cf.NewWithAPIToken(m.APIToken)
	if err != nil {
		// problem wth API token
		return false
	}
	m.api = api

	m.api.SetAuthType(cf.AuthToken)

	return true
}

// GetHostRecord will get details of dns record for a given host
func (m *DNSManager) GetHostRecord(hostname string) (*core.HostInfo, error) {

	//
	baseDomain, err := getBaseDomain(hostname)
	if err != nil {
		return nil, err
	}

	// get zone id
	zoneID, err := m.api.ZoneIDByName(baseDomain)
	if err != nil {
		// check for 'Zone could not be found' means not hosted by CF
		// vs other errors which mean technical issue
		return nil, err
	}

	// get a specific record
	// desiredRec := cf.DNSRecord{Name: hostname, Type: "A"}
	recs, _, err := m.api.ListDNSRecords(context.Background(), cf.ZoneIdentifier(zoneID), cf.ListDNSRecordsParams{Name: hostname, Type: "A"})
	if err != nil {
		// problem making request
		return nil, err
	}

	// desired record does not exist
	if len(recs) == 0 {
		return nil, errors.New("record does not exist")
	}

	// found a record
	hostInfo := core.HostInfo{Name: recs[0].Name, IP: recs[0].Content, ID: recs[0].ID}

	return &hostInfo, nil
}

// AddHostRecord will get details of dns record for a given host
func (m *DNSManager) AddHostRecord(hostInfo core.HostInfo) error {

	// get zone id
	zoneID, err := m.getZoneID(hostInfo.Name)
	if err != nil {
		return err
	}

	// need to create the record
	var newRec cf.DNSRecord
	newRec.Name = hostInfo.Name
	newRec.Content = hostInfo.IP
	newRec.Type = "A"

	dnsResponse, err := m.api.CreateDNSRecord(context.Background(), cf.ZoneIdentifier(zoneID), cf.CreateDNSRecordParams{
		Type:    newRec.Type,
		Name:    newRec.Name,
		Content: newRec.Content,
	})
	if err != nil {
		log.Error("Failed to create DNS record:", err)
		return err
	}
	log.Debug(dnsResponse)

	return nil
}

// UpdateHostRecord will update dns record for a given host
func (m *DNSManager) UpdateHostRecord(hostInfo core.HostInfo) error {

	// get zone id
	zoneID, err := m.getZoneID(hostInfo.Name)
	if err != nil {
		return err
	}

	// update the DNS record
	updateParams := cf.UpdateDNSRecordParams{
		Type:    "A",
		Name:    hostInfo.Name,
		Content: hostInfo.IP,
	}
	_, err = m.api.UpdateDNSRecord(context.Background(), cf.ZoneIdentifier(zoneID), updateParams)
	if err != nil {
		log.Error("Failed to update DNS record:", err)
		return err
	}

	return nil
}

// DeleteHostRecord will delete the given dns record at a provider
func (m *DNSManager) DeleteHostRecord(hostInfo core.HostInfo) error {

	// get zone id
	zoneID, err := m.getZoneID(hostInfo.Name)
	if err != nil {
		return err
	}
	// ready to delete the record
	err = m.api.DeleteDNSRecord(context.Background(), cf.ZoneIdentifier(zoneID), hostInfo.ID)
	if err != nil {
		return err
	}

	return nil
}

// getZoneID will return the cloudflare zone id to be used to access info about a given hostname
func (m *DNSManager) getZoneID(hostname string) (ID string, err error) {

	// figure out root domain
	baseDomain, err := getBaseDomain(hostname)
	if err != nil {
		return "0", err
	}

	// get zone id
	zoneID, err := m.api.ZoneIDByName(baseDomain)
	if err != nil {
		// check for 'Zone could not be found' means not hosted by CF
		// vs other errors which mean technical issue
		return "0", err
	}

	return zoneID, nil
}

// getBaseDomain will take a hostname and return the root domain
func getBaseDomain(hostname string) (baseDomain string, err error) {

	// divide hostname into its parts
	parts := strings.Split(hostname, ".")
	numParts := len(parts)
	if numParts < 2 {
		return "", errors.New("invalate hostname")
	}

	// reassemble base domain (domain + TLDN)
	baseDomain = parts[numParts-2] + "." + parts[numParts-1]

	return baseDomain, nil

}
