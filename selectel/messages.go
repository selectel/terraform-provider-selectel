package selectel

import "fmt"

func msgCreate(object string, options interface{}) string {
	return fmt.Sprintf("[DEBUG] Creating %s with options: %+v", object, options)
}

func msgGet(object, id string) string {
	return fmt.Sprintf("[DEBUG] Getting %s '%s'", object, id)
}

func msgUpdate(object, id string, options interface{}) string {
	return fmt.Sprintf("[DEBUG] Updating %s '%s' with options: %+v", object, id, options)
}

func msgDelete(object, id string) string {
	return fmt.Sprintf("[DEBUG] Deleting %s '%s'", object, id)
}

func msgImport(object, id string) string {
	return fmt.Sprintf("[DEBUG] Importing %s '%s'", object, id)
}
