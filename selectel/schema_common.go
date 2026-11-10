package selectel

import (
	"fmt"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	projectIDDescription = "Unique identifier of the associated project, for example, `a07abc12310546f1b9291ab3013a7d75`."

	projectIDFromResource = "Retrieved from the [selectel_vpc_project_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_project_v2) resource."

	projectIDFromControlPanel = "To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ project name ⟶ copy the ID of the required project."

	projectIDLearnMore = "Learn more about [Projects](https://docs.selectel.ru/en/control-panel-actions/projects/about-projects/)."
	regionLearnMore    = "Learn more about available pools in [Product availability by location](https://docs.selectel.ru/en/infrastructure/product-availability-by-location/)."

	upgradeGuideNote = "If you are upgrading from a version earlier than 9.1.0, follow the [upgrade guide](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_9_1) to migrate your resources from the OpenStack provider."

	exampleRegion     = "ru-6"
	exampleResourceID = "b311ce58-2658-46b5-b733-7a0f418703f2"
)

type resourceDocs struct {
	Name string
}

func (r resourceDocs) projectIDResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: projectIDDescription + " " + projectIDFromResource + " " + projectIDLearnMore,
	}
}

func (r resourceDocs) projectIDIdentitySchema() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       projectIDDescription + " " + projectIDFromControlPanel + " " + projectIDLearnMore,
	}
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

func (r resourceDocs) idIdentitySchema(exampleID, controlPanelHint string) *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       joinSentences(fmt.Sprintf("%s, for example, `%s`.", r.idDescription(), exampleID), controlPanelHint),
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
		Description:       joinSentences(r.regionDescription(), controlPanelHint, regionLearnMore),
	}
}

func (r resourceDocs) identitySchemaByID(exampleID, idHint, regionHint string) map[string]*schema.Schema {
	return r.identitySchema(regionHint, map[string]*schema.Schema{
		"id": r.idIdentitySchema(exampleID, idHint),
	})
}

func (r resourceDocs) identitySchema(regionHint string, attrs map[string]*schema.Schema) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"project_id": r.projectIDIdentitySchema(),
		"region":     r.regionIdentitySchema(regionHint),
	}
	maps.Copy(result, attrs)

	return result
}

func identityAttrSchema(parts ...string) *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		RequiredForImport: true,
		Description:       joinSentences(parts...),
	}
}

func (r resourceDocs) withDocsHints(s map[string]*schema.Schema) map[string]*schema.Schema {
	for name, attr := range s {
		for _, hint := range dependencyHints(name, attr) {
			attr.Description += " " + hint
		}
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

func dependencyHints(name string, attr *schema.Schema) []string {
	var hints []string

	if others := without(attr.ConflictsWith, name); len(others) > 0 {
		hints = append(hints, fmt.Sprintf("Conflicts with %s.", joinCode(others)))
	}
	if group := without(attr.ExactlyOneOf, name); len(group) > 0 {
		if len(group) == 1 {
			hints = append(hints, fmt.Sprintf("Specify either `%s` or %s.", name, joinCode(group)))
		} else {
			hints = append(hints, fmt.Sprintf("Exactly one of %s must be specified.", joinCode(append([]string{name}, group...))))
		}
	}
	if group := without(attr.AtLeastOneOf, name); len(group) > 0 {
		hints = append(hints, fmt.Sprintf("At least one of %s must be specified.", joinCode(append([]string{name}, group...))))
	}
	if others := without(attr.RequiredWith, name); len(others) > 0 {
		hints = append(hints, fmt.Sprintf("Requires %s.", joinCode(others)))
	}

	return hints
}

func without(keys []string, name string) []string {
	var rest []string
	for _, k := range keys {
		if k != name {
			rest = append(rest, k)
		}
	}

	return rest
}

func joinCode(keys []string) string {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = "`" + k + "`"
	}
	switch len(quoted) {
	case 1:
		return quoted[0]
	case 2:
		return quoted[0] + " and " + quoted[1]
	default:
		return strings.Join(quoted[:len(quoted)-1], ", ") + ", and " + quoted[len(quoted)-1]
	}
}

type dataSourceDocs struct {
	Name string
}

func (d dataSourceDocs) projectIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: projectIDDescription + " " + projectIDFromResource + " " + projectIDLearnMore,
	}
}

func (d dataSourceDocs) idSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: fmt.Sprintf("Unique identifier of the %s.", d.Name),
	}
}

func (d dataSourceDocs) regionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("Pool where the %s is located, for example, `%s`. %s", d.Name, exampleRegion, regionLearnMore),
	}
}

func joinSentences(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}

	return strings.Join(nonEmpty, " ")
}
