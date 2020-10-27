package linode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eezhee/eezhee/pkg/core"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func strCopy(i string) *string {
	return &i
}

// Manager handles interactions with DigitalOcean API
type Manager struct {
	APIToken string
	api      linodego.Client
}

// NewManager creates a manage object & inits it
func NewManager(providerAPIToken string) (m *Manager) {

	if len(providerAPIToken) == 0 {
		fmt.Println("no linode api token set")
		return nil
	}

	manager := new(Manager)
	manager.APIToken = providerAPIToken
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: manager.APIToken})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	manager.api = linodego.NewClient(oauth2Client)

	return manager
}

// IsSSHKeyUploaded checks if ssh key already uploaded to DigitalOcean
func (m *Manager) IsSSHKeyUploaded(fingerprint string) (bool, error) {

	// ctx := context.TODO()
	return true, nil
}

// SelectClosestRegion will check all DO regions to find the closest
func (m *Manager) SelectClosestRegion() (closestRegion string, err error) {
	return "", nil
}

// GetVMInfo will get details of a VM
func (m *Manager) GetVMInfo(vmID int) (vmInfo core.VMInfo, err error) {

	testProfile := false
	if testProfile {
		profile, err := m.api.GetProfile(context.Background())
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(profile.AuthorizedKeys)
	}

	testImages := false
	if testImages {
		// get list of VM types (~27 variations)
		images, err := m.api.ListImages(context.Background(), nil)
		for _, image := range images {
			fmt.Println("image:", image.Label, " ", image.ID, " ", image.Description)
		}

		desiredImage := images[0].ID
		image, err := m.api.GetImage(context.Background(), desiredImage)
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(image.Type)
	}

	testTypes := true
	if testTypes {
		// get list of VM types (~27 variations)
		vmTypes, err := m.api.ListTypes(context.Background(), nil)
		for _, vmInfo := range vmTypes {
			fmt.Println("type:", vmInfo.ID, " ", vmInfo.Label)
		}

		desiredInstance := vmTypes[0].ID
		instanceType, err := m.api.GetType(context.Background(), desiredInstance)
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(instanceType.ID)
	}

	testTags := false
	if testTags {

		tagExists := false

		// get list of existing tags
		tags, err := m.api.ListTags(context.Background(), nil)
		if err != nil {
			return vmInfo, err
		}

		// does our tag already exist?
		for _, tag := range tags {
			fmt.Println(tag.Label)
			if strings.Compare(tag.Label, "eezhee") == 0 {
				tagExists = true
			}
		}

		if tagExists {
			err = m.api.DeleteTag(context.Background(), "eezhee")
			if err != nil {
				return vmInfo, err
			}
			tagExists = false
		}
		// if not, need to add
		if !tagExists {
			// create the eezhee tag so we can use when we create objects
			tagCreateOpts := linodego.TagCreateOptions{
				Label: "eezhee",
			}
			newTag, err := m.api.CreateTag(context.Background(), tagCreateOpts)
			if err != nil {
				return vmInfo, err
			}
			fmt.Println("tag:", newTag.Label)
		}

	}

	testRegions := false
	if testRegions {
		regions, _ := m.api.ListRegions(context.Background(), nil)
		for _, region := range regions {
			fmt.Println(region.Country, " ", region.ID)
		}
		fmt.Println("")

	}

	testDomains := false
	if testDomains {
		// createDomainOpts := linodego.DomainCreateOptions{
		// 	Domain:   "testing-eezhee.com",
		// 	Type:     linodego.DomainTypeMaster,
		// 	SOAEmail: "support@eezhee.com",
		// }
		// createDomainOpts.Tags = append(createDomainOpts.Tags, "eezhee")
		// newDomain, err := m.api.CreateDomain(context.Background(), createDomainOpts)
		// if err != nil {
		// 	return vmInfo, err
		// }
		// fmt.Println(newDomain.ID)

		// does the domain that we care about exist?
		targetDomain := "testing-eezhee.com"
		domainExists := false

		// list any domains linode is hosting
		domains, err := m.api.ListDomains(context.Background(), nil)
		for _, domain := range domains {
			if strings.Compare(targetDomain, domain.Domain) == 0 {
				domainExists = true
				break
			}
			fmt.Println(domain.ID, " ", domain.Domain)
		}
		if !domainExists {
			// since don't have domain hosted in linode, can't add any records
			return vmInfo, err
		}

		// get details
		// this really isn't needed
		domain := domains[0]
		domainDetails, err := m.api.GetDomain(context.Background(), domain.ID)
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(domainDetails.Domain)

		// records
		domainRecordExists := false

		// see if A record already exists for host
		filter, _ := json.Marshal(map[string]interface{}{"name": "www"})
		listOpts := linodego.NewListOptions(0, string(filter))
		records, err := m.api.ListDomainRecords(context.Background(), domainDetails.ID, listOpts)
		for _, record := range records {
			// type = A and name = www
			if record.Type == "A" && (strings.Compare("www", record.Name)) == 0 {
				domainRecordExists = true
				break
			}

			fmt.Println(record.Name)
		}

		if domainRecordExists {
			// update the record
			recUpdateOpts := linodego.DomainRecordUpdateOptions{
				Target: "192.168.1.1",
			}
			domainRecordID := records[0].ID
			updatedRecord, err := m.api.UpdateDomainRecord(context.Background(), domainDetails.ID, domainRecordID, recUpdateOpts)
			if err != nil {
				return vmInfo, nil
			}
			fmt.Println(updatedRecord.Target)
		} else {
			// create the record
			recCreateOpts := linodego.DomainRecordCreateOptions{
				Type: linodego.RecordTypeA,
				Name: "www",
				// TTLSec: 60,
				Target: "127.0.0.1",
			}
			// recCreateOpts.Tag = strCopy("eezhee") // can't assign a constant
			newRecord, err := m.api.CreateDomainRecord(context.Background(), domainDetails.ID, recCreateOpts)
			if err != nil {
				return vmInfo, err
			}
			fmt.Println(newRecord.ID)
		}

		// err = m.api.DeleteDomainRecord(context.Background(), domainDetails.ID, domainDetails.ID)
		// if err != nil {
		// 	// continue with domain deletion
		// }

		// // remove domains
		// err = m.api.DeleteDomain(context.Background(), domainDetails.ID)
		// if err != nil {
		// 	return vmInfo, err
		// }
	}

	testSSHKeys := false
	if testSSHKeys {

		// NOTE: linode will allow multiple ssh keys with the same name and same key value

		// find the image we want to use
		// keyExists := false
		keys, err := m.api.ListSSHKeys(context.Background(), nil)
		for _, key := range keys {

			// get SSHKey, get fingerprint of key and see if matches id_rsa.pub otherwise create it
			// fingerprint := key.SSHKey
			fmt.Println(key.ID, " ", key.SSHKey)
		}

		sshOptions := linodego.SSHKeyCreateOptions{
			Label:  "athirnuaimi2-eezhee",
			SSHKey: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAocQ68fyqU/QZJYrpGrM+tkJDfUPefFPa2Qc+C2BHom3gysv8vqwmFgdVs6Z75rPkUNitpIxUYGPovJbG5pFE6qRNxK3ZHxbk1TSlFBcL8w7jd/jt4IuHwslO4R+hxLG0vzGVFpKSKjAM6yac+q8wOtFU7pKpmrGx9oyClrVQb4mSbCdDazf7/uzXpKMg5mgONbjT6AWSpos2cUDH+VNAQKEnFxKWYjEddCqJnN2kIvtvJUeVhaxYjSVgtiJ7/e0KboDBKRRtO+b4v2TmWmGoRhrPqMo3GazU9aSOAEOMrl3SrxkjmH+eRCUA+1zdvwes8ncVK36FNXzFJ7CxGEAHrw== athir@nuaimi.com",
		}
		newKey, err := m.api.CreateSSHKey(context.Background(), sshOptions)
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(newKey)

	}

	testInstances := true
	if testInstances {
		instances, err := m.api.ListInstances(context.Background(), nil)
		for _, instance := range instances {
			fmt.Println(instance.ID, " ", instance.Label, " ", instance.IPv4)
		}

		createOptions := linodego.InstanceCreateOptions{
			Region:   "ca-central",
			Type:     "g6-nanode-1",
			Label:    "test-001",
			Image:    "linode/ubuntu20.04",
			RootPass: "why-is-this-needed-2jkaf832",
		}
		// createOptions.AuthorizedUsers = append(createOptions.AuthorizedUsers, "anuaimi")
		createOptions.AuthorizedKeys = append(createOptions.AuthorizedKeys, "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAocQ68fyqU/QZJYrpGrM+tkJDfUPefFPa2Qc+C2BHom3gysv8vqwmFgdVs6Z75rPkUNitpIxUYGPovJbG5pFE6qRNxK3ZHxbk1TSlFBcL8w7jd/jt4IuHwslO4R+hxLG0vzGVFpKSKjAM6yac+q8wOtFU7pKpmrGx9oyClrVQb4mSbCdDazf7/uzXpKMg5mgONbjT6AWSpos2cUDH+VNAQKEnFxKWYjEddCqJnN2kIvtvJUeVhaxYjSVgtiJ7/e0KboDBKRRtO+b4v2TmWmGoRhrPqMo3GazU9aSOAEOMrl3SrxkjmH+eRCUA+1zdvwes8ncVK36FNXzFJ7CxGEAHrw== athir@nuaimi.com")
		createOptions.Tags = append(createOptions.Tags, "eezhee")
		newInstance, err := m.api.CreateInstance(context.Background(), createOptions)
		if err != nil {
			return vmInfo, err
		}
		fmt.Println(newInstance.ID)
		fmt.Println(newInstance.Status)

		// see if vm ready
		status := newInstance.Status
		for status != linodego.InstanceRunning {

			// wait a bit
			time.Sleep(2 * time.Second)

			instanceInfo, err := m.api.GetInstance(context.Background(), newInstance.ID)
			if err != nil {
				return vmInfo, err
			}

			status = instanceInfo.Status
		}

		// now resize it
		resizeOptions := linodego.InstanceResizeOptions{
			Type: "g6-standard-1",
		}
		err = m.api.ResizeInstance(context.Background(), newInstance.ID, resizeOptions)
		if err != nil {
			return vmInfo, err
		}

	}

	// create an instance

	// ssh keys

	// locations

	return vmInfo, nil
}
