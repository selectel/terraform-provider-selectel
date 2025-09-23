package selectel

import (
	"fmt"
)

var ErrNotFound = fmt.Errorf("not found")

func errParsingPrefixLength(object, id string, err error) string {
	return fmt.Sprintf("[DEBUG] can't parse prefix length from %s '%s' CIDR: %s", object, id, err)
}

func errSettingComplexAttr(attr string, err error) string {
	return fmt.Sprintf("[DEBUG] error setting %s: %s", attr, err)
}

func errParseID(object, id string) error {
	return fmt.Errorf("unable to parse %s ID: '%s'", object, id)
}

func errParseProjectV2Quotas(err error) error {
	return fmt.Errorf("got error parsing quotas: %s", err)
}

func errParseCrossRegionSubnetV2Regions(err error) error {
	return fmt.Errorf("got error parsing regions: %s", err)
}

func errParseCrossRegionSubnetV2ProjectID(err error) error {
	return fmt.Errorf("got error parsing project ID: %s", err)
}

func errParseMKSNodegroupV1ID(nodegroupID string) error {
	return fmt.Errorf("got error parsing nodegroup ID: %s", nodegroupID)
}

func errParseDomainsDomainV1ID(domainID string) error {
	return fmt.Errorf("got error parsing domain ID: %s", domainID)
}

func errParseDomainsRecordV1ID(domainID string) error {
	return fmt.Errorf("got error parsing record ID: %s", domainID)
}

func errParseDomainsDomainRecordV1IDsPair(id string) error {
	return fmt.Errorf("got error parsing domain/record IDs pair: %s", id)
}

func errSearchingProjectRole(projectID string, err error) error {
	return fmt.Errorf("can't find role for project '%s': %s", projectID, err)
}

func errSearchingKeypair(keypairName string, err error) error {
	return fmt.Errorf("can't find keypair '%s': %s", keypairName, err)
}

func errReadFromResponse(object string) error {
	return fmt.Errorf("can't get %s from the response", object)
}

func errCreatingObject(object string, err error) error {
	return fmt.Errorf("error creating %s: %s", object, err)
}

func errUpdatingObject(object, id string, err error) error {
	return fmt.Errorf("error updating %s '%s': %s", object, id, err)
}

func errGettingObject(object, id string, err error) error {
	return fmt.Errorf("error getting %s '%s': %s", object, id, err)
}

func errDeletingObject(object, id string, err error) error {
	return fmt.Errorf("error deleting %s '%s': %s", object, id, err)
}

func errResourceDeprecated(resource string) error {
	return fmt.Errorf("%s resource has been deprecated", resource)
}

func errKubeVersionIsInvalidFmt(kubeVersion string, errMsg string) error {
	return fmt.Errorf("kube version %s is invalid: %s", kubeVersion, errMsg)
}

func errGettingObjects(object string, err error) error {
	return fmt.Errorf("error getting %s: %s", object, err)
}

func errParseDatastoreV1Flavor(err error) error {
	return fmt.Errorf("got error parsing flavor: %s", err)
}

func errParseDatastoreV1Pooler(err error) error {
	return fmt.Errorf("got error parsing pooler opts: %s", err)
}

func errParseDatastoreV1Firewall(err error) error {
	return fmt.Errorf("got error parsing firewall opts: %s", err)
}

func errParseDatastoreV1Resize(err error) error {
	return fmt.Errorf("got error parsing resize opts: %s", err)
}

func errParseDatastoreV1Restore(err error) error {
	return fmt.Errorf("got error parsing restore opts: %s", err)
}

func errParseDatastoreV1FloatingIPs(err error) error {
	return fmt.Errorf("got error parsing floating IPs opts: %s", err)
}
