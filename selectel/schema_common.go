package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func projectIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: "Unique identifier of the associated project. " +
			"Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource. " +
			"Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/).",
	}
}

func regionSchema(objectName, exampleRegion string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: fmt.Sprintf("Pool where the %s is located, for example, `%s`. "+
			"Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/).",
			objectName, exampleRegion),
	}
}

func withDocsHints(objectName string, s map[string]*schema.Schema) map[string]*schema.Schema {
	for _, attr := range s {
		if attr.ForceNew {
			attr.Description += fmt.Sprintf(" Changing this creates a new %s.", objectName)
		}
		if attr.Default != nil {
			if attr.Default == "" {
				attr.Description += " The default value is an empty string."
			} else {
				attr.Description += fmt.Sprintf(" The default value is `%v`.", attr.Default)
			}
		}
		if nested, ok := attr.Elem.(*schema.Resource); ok {
			withDocsHints(objectName, nested.Schema)
		}
	}

	return s
}
