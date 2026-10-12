package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	projectIDDescription      = "Unique identifier of the associated project, for example, `a07abc12310546f1b9291ab3013a7d75`."
	projectIDFromResource     = "Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource."
	projectIDFromControlPanel = "To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ project name ⟶ copy the ID of the required project."
	projectIDLearnMore        = "Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/)."

	regionLearnMore = "Learn more about available pools in [Product availability by location](https://docs.selectel.ru/en/infrastructure/product-availability-by-location/)."

	exampleRegion     = "ru-6"
	exampleResourceID = "b311ce58-2658-46b5-b733-7a0f418703f2"
)

type resourceDocs struct {
	Name string
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
		Description:       fmt.Sprintf("%s, for example, `%s`. %s", r.idDescription(), exampleResourceID, controlPanelHint),
	}
}

func (r resourceDocs) regionDescription() string {
	return fmt.Sprintf("Pool where the %s is located, for example, `%s`.", r.Name, exampleRegion)
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
			attr.Description += defaultHint(attr.Default)
		}
		if nested, ok := attr.Elem.(*schema.Resource); ok {
			r.withDocsHints(nested.Schema)
		}
	}

	return s
}

func defaultHint(v any) string {
	if v == "" {
		return " The default value is an empty string."
	}

	return fmt.Sprintf(" The default value is `%v`.", v)
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
