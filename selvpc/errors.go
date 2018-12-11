package selvpc

import "fmt"

func errParseProjectV2Quotas(err error) error {
	return fmt.Errorf("got error parsing quotas: %s", err)
}

func errSearchingProjectRole(projectID string, err error) error {
	return fmt.Errorf("can't find role for project '%s': %s", projectID, err)
}

func errCreatingObject(object string, err error) error {
	return fmt.Errorf("[DEBUG] error creating %s: %s", object, err)
}

func errUpdatingObject(object, id string, err error) error {
	return fmt.Errorf("[DEBUG] error updating %s '%s': %s", object, id, err)
}

func errGettingObject(object, id string, err error) error {
	return fmt.Errorf("[DEBUG] error getting %s '%s': %s", object, id, err)
}

func errDeletingObject(object, id string, err error) error {
	return fmt.Errorf("[DEBUG] error deleting %s '%s': %s", object, id, err)
}
