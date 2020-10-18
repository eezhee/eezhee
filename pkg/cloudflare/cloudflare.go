package cloudflare

import (
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
	cf "github.com/cloudflare/cloudflare-go"
)

// Test allows us to get familiar with the Cloudflare API
func Test() bool {

	// make sure we have an API token, else return

	api, err := cf.NewWithAPIToken("g3rUkFXNCb6dTvYt-JyZ1Eoh27TFNnjYCFK58mSp")
	if err != nil {
		//
		return false
	}

	api.SetAuthType(cf.AuthToken)

	// u, err := api.UserDetails()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // Print user details
	// fmt.Println(u)

	// see if domain hosted at Cloudflare
	// can also use ListZones()

	id, err := api.ZoneIDByName("rndguy.ca")
	if err != nil {
		// check for 'Zone could not be found' means not hosted by CF
		// vs other errors which mean technical issue
		log.Fatal(err)
		return false
	}

	// Fetch zone details
	// note: not sure we need this
	zone, err := api.ZoneDetails(id)
	if err != nil {
		log.Fatal(err)
		return false
	}
	// Print zone details
	fmt.Println(zone)

	// get all the records
	recs, err := api.DNSRecords(id, cloudflare.DNSRecord{})
	if err != nil {
		log.Fatal(err)
		return false
	}

	for _, r := range recs {
		fmt.Printf("%s: %s\n", r.Name, r.Content)
	}

	// get a specific record
	desiredRec := cloudflare.DNSRecord{Name: "cluster1.k8s.rndguy.ca"}
	recs, err = api.DNSRecords(id, desiredRec)
	if err != nil {
		log.Fatal(err)
		// A rec does not exists
		// lets create it
		return false
	}

	// record exists, do we need to update it
	newRec := cloudflare.DNSRecord{Name: "newcluster.k8s.rndguy.ca", Type: "A"}
	recs, err = api.DNSRecords(id, newRec)
	if err != nil {
		log.Fatal(err)
		// A rec does not exists
		// lets create it
		return false
	}

	newRec.Content = "68.131.1.1"

	if len(recs) > 0 {

		// delete the record (so we can update it)
		err = api.UpdateDNSRecord(id, recs[0].ID, newRec)
		// err = api.DeleteDNSRecord(id, recs[0].ID)
		if err != nil {
			log.Fatal(err)
			return false
		}

	} else {
		// need to create the record
		dnsResponse, err := api.CreateDNSRecord(id, newRec)
		if err != nil {
			log.Fatal(err)
			return false
		}
		fmt.Println(dnsResponse)

	}

	return true
}
