package selectel

import (
	"strconv"
	"strings"
	"time"
)

const (
	domainsV1DefaultRetryWaitMin = time.Second //nolint:revive
	domainsV1DefaultRetryWaitMax = 5 * time.Second
	domainsV1DefaultRetry        = 5
)

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
