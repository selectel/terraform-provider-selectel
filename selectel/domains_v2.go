package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

var ErrProjectIDNotSetupForDNSV2 = errors.New("env variable SEL_PROJECT_ID or variable project_id must be set for the dns v2")

func getDomainsV2Client(meta interface{}) (domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, ErrProjectIDNotSetupForDNSV2
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for domains v2: %w", err)
	}

	httpClient := &http.Client{}
	userAgent := "terraform-provider-selectel"
	defaultApiURL := "https://api.selectel.ru/domains/v2"
	hdrs := http.Header{}
	hdrs.Add("X-Auth-Token", selvpcClient.GetXAuthToken())
	hdrs.Add("User-Agent", userAgent)
	domainsClient := domainsV2.NewClient(defaultApiURL, httpClient, hdrs)

	return domainsClient, nil
}

func getZoneByName(ctx context.Context, client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], zoneName string) (*domainsV2.Zone, error) {
	optsForSearchZone := map[string]string{
		"filter": zoneName,
		"limit":  "1000",
		"offset": "0",
	}
	r, err := regexp.Compile(fmt.Sprintf("^%s.?", zoneName))
	if err != nil {
		return nil, err
	}

	for {
		zones, err := client.ListZones(ctx, &optsForSearchZone)
		if err != nil {
			return nil, err
		}

		for _, zone := range zones.GetItems() {
			if r.MatchString(zone.Name) {
				return zone, nil
			}
		}
		optsForSearchZone["offset"] = strconv.Itoa(zones.GetNextOffset())
		if zones.GetNextOffset() == 0 {
			break
		}
	}

	return nil, errGettingObject(objectZone, zoneName, ErrZoneNotFound)
}

func getRrsetByNameAndType(ctx context.Context, client domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], zoneID, rrsetName, rrsetType string) (*domainsV2.RRSet, error) {
	optsForSearchRrset := map[string]string{
		"name":        rrsetName,
		"rrset_types": rrsetType,
		"limit":       "1000",
		"offset":      "0",
	}

	r, err := regexp.Compile(fmt.Sprintf("^%s.?", rrsetName))
	if err != nil {
		return nil, errGettingObject(objectRrset, rrsetName, err)
	}

	for {
		rrsets, err := client.ListRRSets(ctx, zoneID, &optsForSearchRrset)
		if err != nil {
			return nil, errGettingObject(objectRrset, rrsetName, err)
		}
		for _, rrset := range rrsets.GetItems() {
			if r.MatchString(rrset.Name) && string(rrset.Type) == rrsetType {
				return rrset, nil
			}
		}
		optsForSearchRrset["offset"] = strconv.Itoa(rrsets.GetNextOffset())
		if rrsets.GetNextOffset() == 0 {
			break
		}
	}

	return nil, errGettingObject(objectRrset, fmt.Sprintf("Name: %s. Type: %s.", rrsetName, rrsetType), ErrRrsetNotFound)
}
