package selectel

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	domainsV1 "github.com/selectel/domains-go/pkg/v1"
)

const (
	domainsV1DefaultRetryWaitMin = time.Second //nolint:revive
	domainsV1DefaultRetryWaitMax = 5 * time.Second
	domainsV1DefaultRetry        = 5
)

func getDomainsClient(meta interface{}) (*domainsV1.ServiceClient, error) {
	config := meta.(*Config)

	selvpcClient, err := config.GetSelVPCClient()
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for domains: %w", err)
	}

	domainsClient := domainsV1.NewDomainsClientV1WithDefaultEndpoint(selvpcClient.GetXAuthToken()).WithOSToken()

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil // Ignore retyablehttp client logs
	retryClient.RetryWaitMin = domainsV1DefaultRetryWaitMin
	retryClient.RetryWaitMax = domainsV1DefaultRetryWaitMax
	retryClient.RetryMax = domainsV1DefaultRetry
	domainsClient.HTTPClient = retryClient.StandardClient()

	return domainsClient, nil
}

const (
	TypeRecordA     string = "A"
	TypeRecordAAAA  string = "AAAA"
	TypeRecordTXT   string = "TXT"
	TypeRecordCNAME string = "CNAME"
	TypeRecordNS    string = "NS"
	TypeRecordSOA   string = "SOA"
	TypeRecordMX    string = "MX"
	TypeRecordSRV   string = "SRV"
	TypeRecordCAA   string = "CAA"
	TypeRecordSSHFP string = "SSHFP"
	TypeRecordALIAS string = "ALIAS"
)

func domainsV1ParseDomainRecordIDsPair(id string) (int, int, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return -1, -1, errParseDomainsDomainRecordV1IDsPair(id)
	}
	if parts[0] == "" || parts[1] == "" {
		return -1, -1, errParseDomainsDomainRecordV1IDsPair(id)
	}

	domainID, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1, -1, errParseDomainsDomainV1ID(parts[0])
	}

	recordID, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, -1, errParseDomainsRecordV1ID(parts[1])
	}

	return domainID, recordID, nil
}

func getIntPtrOrNil(v interface{}) *int {
	if v == nil {
		return nil
	}

	return intPtr(v.(int))
}

func intPtr(v int) *int {
	return &v
}
