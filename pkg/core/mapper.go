package core

import (
	"encoding/json"
)

type Location struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	State   string `json:"state"`
	City    string `json:"city"`
}

type ProviderMapping struct {
	Image   string              `json:"image"`
	VMSizes map[string]string   `json:"sizes"`
	Regions map[string]Location `json:"regions"`
}

// ParseProviderMappings reads in the JSON file that maps a provider services to Eezhee format
func ParseProviderMappings(providerMappings []byte) (pm ProviderMapping, err error) {
	err = json.Unmarshal([]byte(providerMappings), &pm)
	return pm, err
}
