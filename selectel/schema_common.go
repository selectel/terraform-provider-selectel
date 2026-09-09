package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	projectIDDescription      = "Unique identifier of the associated project."
	projectIDFromResource     = "Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource."
	projectIDFromControlPanel = "To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ project name ⟶ copy the ID of the required project."
	projectIDLearnMore        = "Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/)."

	regionLearnMore = "Learn more about available pools in the [Availability matrix](https://docs.selectel.ru/en/control-panel-actions/availability-matrix/)."
)

type resourceDocs struct {
	Name          string
	ExampleRegion string
	ExampleID     string
}

func (r resourceDocs) idDescription() string {
	return fmt.Sprintf("Unique identifier of the %s", r.Name)
}

func (r resourceDocs) idResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: r.idDescription() + ".",
	}
}

func (r resourceDocs) idIdentitySchema(controlPanelHint string) *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       fmt.Sprintf("%s, for example, `%s`. %s", r.idDescription(), r.ExampleID, controlPanelHint),
	}
}

func (r resourceDocs) regionDescription() string {
	return fmt.Sprintf("Pool where the %s is located, for example, `%s`.", r.Name, r.ExampleRegion)
}

func (r resourceDocs) regionResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: r.regionDescription() + " " + regionLearnMore,
	}
}

func (r resourceDocs) regionIdentitySchema(controlPanelHint string) *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       r.regionDescription() + " " + controlPanelHint + " " + regionLearnMore,
	}
}

func (r resourceDocs) withDocsHints(s map[string]*schema.Schema) map[string]*schema.Schema {
	for _, attr := range s {
		if attr.ForceNew {
			attr.Description += fmt.Sprintf(" Changing this creates a new %s.", r.Name)
		}
		if attr.Default != nil {
			if attr.Default == "" {
				attr.Description += " The default value is an empty string."
			} else {
				attr.Description += fmt.Sprintf(" The default value is `%v`.", attr.Default)
			}
		}
		if nested, ok := attr.Elem.(*schema.Resource); ok {
			r.withDocsHints(nested.Schema)
		}
	}

	return s
}

func projectIDResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: projectIDDescription + " " + projectIDFromResource + " " + projectIDLearnMore,
	}
}

func projectIDIdentitySchema() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       projectIDDescription + " " + projectIDFromControlPanel + " " + projectIDLearnMore,
	}
}
