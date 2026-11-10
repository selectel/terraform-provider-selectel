package schemadoc

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringRule struct {
	validator.String
	Facts
}

type int64Rule struct {
	validator.Int64
	Facts
}

type setRule struct {
	validator.Set
	Facts
}

type mapRule struct {
	validator.Map
	Facts
}

type stringDefault struct {
	defaults.String
	Facts
}

type boolDefault struct {
	defaults.Bool
	Facts
}

type int64Default struct {
	defaults.Int64
	Facts
}

type stringReplacement struct {
	planmodifier.String
	Facts
}

// LengthAtMost counts characters, not UTF-8 bytes.
func LengthAtMost(length int) validator.String {
	return stringRule{stringvalidator.UTF8LengthAtMost(length), Facts{
		Constraint: fmt.Sprintf("The maximum length is `%d` %s.", length, characters(length)),
	}}
}

// LengthAtLeast counts characters, not UTF-8 bytes.
func LengthAtLeast(length int) validator.String {
	return stringRule{stringvalidator.UTF8LengthAtLeast(length), Facts{
		Constraint: fmt.Sprintf("The minimum length is `%d` %s.", length, characters(length)),
	}}
}

func Between(minimum, maximum int64) validator.Int64 {
	return int64Rule{int64validator.Between(minimum, maximum), Facts{
		Constraint: fmt.Sprintf("The available range is from `%d` to `%d`.", minimum, maximum),
	}}
}

func AtLeast(minimum int64) validator.Int64 {
	return int64Rule{int64validator.AtLeast(minimum), Facts{
		Constraint: fmt.Sprintf("The minimum value is `%d`.", minimum),
	}}
}

// relationValidator is a HashiCorp schema validator that works for every
// attribute kind, such as the result of stringvalidator.ConflictsWith.
type relationValidator interface {
	validator.Bool
	validator.Int64
	validator.Map
	validator.String
}

// Relation is a dependency between attributes. One value fits the Validators
// of any attribute kind.
type Relation struct {
	relationValidator
	Facts
}

// ConflictsWith rejects the configuration when any of the named root
// attributes is set together with this one.
func ConflictsWith(attributes ...string) Relation {
	return Relation{
		stringvalidator.ConflictsWith(rootPaths(attributes)...).(relationValidator),
		Facts{Dependency: "This argument conflicts with " + CodeList(attributes) + "."},
	}
}

// ExactlyOneOf requires exactly one of the named root attributes. List the
// whole group, including the attribute that has the validator, so every
// member gets the same sentence.
func ExactlyOneOf(attributes ...string) Relation {
	return Relation{
		stringvalidator.ExactlyOneOf(rootPaths(attributes)...).(relationValidator),
		Facts{Dependency: "Set exactly one of " + CodeList(attributes) + "."},
	}
}

// AlsoRequires rejects the configuration when this attribute is set without all
// of the named root attributes.
func AlsoRequires(attributes ...string) Relation {
	return Relation{
		stringvalidator.AlsoRequires(rootPaths(attributes)...).(relationValidator),
		Facts{Dependency: "This argument requires " + CodeList(attributes) + "."},
	}
}

func rootPaths(attributes []string) []path.Expression {
	paths := make([]path.Expression, len(attributes))
	for i, name := range attributes {
		paths[i] = path.MatchRoot(name)
	}

	return paths
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// UUID accepts the lowercase form in which the API returns identifiers, so a
// reference that the API reports back never differs from the configured value.
func UUID() validator.String {
	return DescribeString(stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"), "The format is a lowercase UUID.")
}

// space is the whitespace that Python str.strip() removes, which is how the
// OpenStack APIs trim names and descriptions: ASCII and Unicode separators,
// including \v and the no-break space.
const space = `\s\x0b\x1c-\x1f\x85\pZ`

var (
	trimmedPattern         = regexp.MustCompile(`^(?s:[^` + space + `](.*[^` + space + `])?)?$`)
	nonEmptyTrimmedPattern = regexp.MustCompile(`^(?s:[^` + space + `](.*[^` + space + `])?)$`)
)

// Trimmed rejects leading and trailing whitespace, which the API would strip
// and make the plan disagree with the stored value. An empty string is allowed.
func Trimmed() validator.String {
	return DescribeString(
		stringvalidator.RegexMatches(trimmedPattern, "must not start or end with whitespace"),
		"The value cannot start or end with whitespace.",
	)
}

// NonEmptyTrimmed is Trimmed that also rejects an empty string.
func NonEmptyTrimmed() validator.String {
	return DescribeString(
		stringvalidator.RegexMatches(nonEmptyTrimmedPattern, "must be nonempty and not start or end with whitespace"),
		"The value cannot be empty or start or end with whitespace.",
	)
}

func StaticString(value string) defaults.String {
	text := fmt.Sprintf("The default value is `%s`.", value)
	if value == "" {
		text = "The default value is an empty string."
	}

	return stringDefault{stringdefault.StaticString(value), Facts{Default: text}}
}

func StaticBool(value bool) defaults.Bool {
	return boolDefault{booldefault.StaticBool(value), Facts{
		Default: fmt.Sprintf("The default value is `%t`.", value),
	}}
}

func StaticInt64(value int64) defaults.Int64 {
	return int64Default{int64default.StaticInt64(value), Facts{
		Default: fmt.Sprintf("The default value is `%d`.", value),
	}}
}

// RequiresReplace has the same behavior as stringplanmodifier.RequiresReplace.
func RequiresReplace() planmodifier.String {
	return stringReplacement{stringplanmodifier.RequiresReplace(), Facts{ReplacesResource: true}}
}

// RequiresReplaceIf has the same behavior as stringplanmodifier.RequiresReplaceIf.
// exception is a sentence about the changes that do not replace the resource.
// The description still ends with the replacement sentence.
func RequiresReplaceIf(replace stringplanmodifier.RequiresReplaceIfFunc, exception string) planmodifier.String {
	return stringReplacement{
		stringplanmodifier.RequiresReplaceIf(replace, exception, exception),
		Facts{ReplacesResource: true, ReplacementException: exception},
	}
}

// DescribeString attaches a service-specific explanation to an existing validator.
// Keep this adapter beside that validator's constructor so its behavior and prose
// are maintained together.
func DescribeString(rule validator.String, description string) validator.String {
	return stringRule{rule, Facts{Constraint: description}}
}

// DescribeSet attaches an explanation to an existing Set validator, as DescribeString does.
func DescribeSet(rule validator.Set, description string) validator.Set {
	return setRule{rule, Facts{Constraint: description}}
}

// DescribeMap attaches an explanation to an existing Map validator, as DescribeString does.
func DescribeMap(rule validator.Map, description string) validator.Map {
	return mapRule{rule, Facts{Constraint: description}}
}

// OneOf uses the same case-sensitive values for validation and documentation.
func OneOf(values ...string) validator.String {
	text := "Available values are " + CodeList(values) + "."
	if len(values) == 1 {
		text = "The only available value is " + CodeList(values) + "."
	}

	return DescribeString(stringvalidator.OneOf(values...), text)
}

func characters(n int) string {
	if n == 1 {
		return "character"
	}

	return "characters"
}
