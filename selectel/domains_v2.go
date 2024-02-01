package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domainsV2 "github.com/selectel/domains-go/pkg/v2"
)

var ErrProjectIDNotSetupForDNSV2 = errors.New("env variable SEL_PROJECT_ID or variable project_id must be set for the dns v2")

func getDomainsV2Client(d *schema.ResourceData, meta interface{}) (domainsV2.DNSClient[domainsV2.Zone, domainsV2.RRSet], error) {
	config := meta.(*Config)
	projectID, err := getProjectIDFromResourceOrConfig(d, config)
	if err != nil {
		return nil, fmt.Errorf("can't get projectID for domains v2: %w", err)
	}
	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for domains v2: %w", err)
	}

	httpClient := &http.Client{}
	userAgent := "terraform-provider-selectel"
	defaultAPIURL := "https://api.selectel.ru/domains/v2"
	hdrs := http.Header{}
	hdrs.Add("X-Auth-Token", selvpcClient.GetXAuthToken())
	hdrs.Add("User-Agent", userAgent)
	domainsClient := domainsV2.NewClient(defaultAPIURL, httpClient, hdrs)

	return domainsClient, nil
}

func getProjectIDFromResourceOrConfig(d *schema.ResourceData, config *Config) (string, error) {
	projectID := config.ProjectID
	if v, ok := d.GetOk("project_id"); ok {
		projectID = v.(string)
	}
	if projectID == "" {
		return "", ErrProjectIDNotSetupForDNSV2
	}

	return projectID, nil
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

func setZoneToResourceData(d *schema.ResourceData, zone *domainsV2.Zone) error {
	d.SetId(zone.ID)
	d.Set("name", zone.Name)
	d.Set("comment", zone.Comment)
	d.Set("created_at", zone.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", zone.UpdatedAt.Format(time.RFC3339))
	d.Set("delegation_checked_at", zone.DelegationCheckedAt.Format(time.RFC3339))
	d.Set("last_check_status", zone.LastCheckStatus)
	d.Set("last_delegated_at", zone.LastDelegatedAt.Format(time.RFC3339))
	d.Set("project_id", strings.ReplaceAll(zone.ProjectID, "-", ""))
	d.Set("disabled", zone.Disabled)

	return nil
}

func setRrsetToResourceData(d *schema.ResourceData, rrset *domainsV2.RRSet) error {
	d.SetId(rrset.ID)
	d.Set("name", rrset.Name)
	d.Set("comment", rrset.Comment)
	d.Set("managed_by", rrset.ManagedBy)
	d.Set("ttl", rrset.TTL)
	d.Set("type", rrset.Type)
	d.Set("zone_id", rrset.ZoneID)
	d.Set("records", generateSetFromRecords(rrset.Records))

	return nil
}

// generateSetFromRecords - generate terraform TypeList from records in rrset.
func generateSetFromRecords(records []domainsV2.RecordItem) []interface{} {
	recordsAsList := []interface{}{}
	for _, record := range records {
		recordsAsList = append(recordsAsList, map[string]interface{}{
			"content":  record.Content,
			"disabled": record.Disabled,
		})
	}

	return recordsAsList
}

// generateRecordsFromSet - generate records for Rrset from terraform TypeList.
func generateRecordsFromSet(recordsSet *schema.Set) []domainsV2.RecordItem {
	records := []domainsV2.RecordItem{}
	for _, recordItem := range recordsSet.List() {
		if record, isOk := recordItem.(map[string]interface{}); isOk {
			records = append(records, domainsV2.RecordItem{
				Content:  record["content"].(string),
				Disabled: record["disabled"].(bool),
			})
		}
	}

	return records
}
