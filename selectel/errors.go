package selectel

import "fmt"

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
