package images

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

var imageTimeouts = schemadoc.Timeouts{
	Create: 30 * time.Minute,
	Delete: 10 * time.Minute,
}

// Formats accepted by the Image service ([image_format] of glance-api.conf).
var (
	containerFormats = []string{"bare", "ovf", "aki", "ari", "ami", "ova", "docker", "compressed"}
	diskFormats      = []string{"raw", "vhd", "vhdx", "vmdk", "vdi", "iso", "ploop", "qcow2", "aki", "ari", "ami"}
)

var (
	aboutImages     = "https://docs.selectel.ru/en/cloud-servers/images/about-images/"
	formatsLink     = schemadoc.Link{Title: "supported formats", URL: aboutImages + "#own-images"}
	accessRolesDocs = "https://docs.selectel.ru/en/cloud-servers/images/manage-access/"
	sharingLink     = schemadoc.Link{
		Title: "image sharing",
		URL:   "https://docs.selectel.ru/en/cloud-servers/images/transfer-image/#configure-access-to-image",
	}
)

// webURL accepts the URLs that the web-download import method can fetch.
func webURL() validator.String {
	return schemadoc.DescribeString(
		stringvalidator.RegexMatches(regexp.MustCompile(`^https?://.+`), "must be an http or https URL"),
		"The URL must start with `http://` or `https://`.",
	)
}

// Limits of the Image service: image_tag_quota and image_property_quota of
// glance-api.conf. The property quota counts the properties that the service
// adds as well.
func tagsRules() []validator.Set {
	return []validator.Set{
		schemadoc.DescribeSet(setvalidator.SizeAtMost(128), "The maximum number of tags is `128`."),
		schemadoc.DescribeSet(
			setvalidator.ValueStringsAre(schemadoc.LengthAtLeast(1), schemadoc.LengthAtMost(255)),
			"Each tag must contain from `1` to `255` characters.",
		),
	}
}

func propertiesLimit() validator.Map {
	return schemadoc.DescribeMap(
		mapvalidator.SizeAtMost(128),
		"The maximum number of properties is `128`, including the properties added by the API.",
	)
}

// propertyKeys applies to every property key. The Image service strips
// whitespace from the parts of a JSON Patch path, so a key with surrounding
// whitespace could be created but never updated or removed.
func propertyKeys() validator.Map {
	return schemadoc.DescribeMap(
		mapvalidator.KeysAre(schemadoc.NonEmptyTrimmed(), schemadoc.LengthAtMost(255)),
		"Each key must contain from `1` to `255` characters and cannot start or end with whitespace.",
	)
}

// baseFields are the image fields that a property key cannot use, with the
// argument that sets the field where there is one.
var baseFields = map[string]string{
	"name": "name", "visibility": "visibility", "protected": "protected", "os_hidden": "hidden",
	"min_disk": "min_disk_gb", "min_ram": "min_ram_mb", "tags": "tags",
	"container_format": "container_format", "disk_format": "disk_format",
	"id": "", "status": "", "checksum": "", "os_hash_algo": "", "os_hash_value": "", "owner": "",
	"size": "", "virtual_size": "", "created_at": "", "updated_at": "", "file": "", "self": "",
	"schema": "", "direct_url": "", "stores": "", "locations": "",
}

// reservedPropertyPrefix marks the properties that the Image service manages.
const reservedPropertyPrefix = "os_glance"

// sortKeys are the image fields that the Image service sorts by.
var sortKeys = []string{"name", "status", "container_format", "disk_format", "size", "id", "created_at", "updated_at"}

func sortOrder() validator.String {
	item := `(` + strings.Join(sortKeys, "|") + `)(:(asc|desc))?`
	return schemadoc.DescribeString(
		stringvalidator.RegexMatches(regexp.MustCompile(`^`+item+`(,`+item+`)*$`), "must be a comma-separated list of <key>[:<direction>] items"),
		"The format is a comma-separated list of `<key>` or `<key>:<direction>` items. Available keys are "+
			schemadoc.CodeList(sortKeys)+". Available directions are `asc` and `desc`. "+
			"An item without a direction is sorted in the `desc` direction.",
	)
}

// regexpValidator accepts the RE2 syntax of the Go regexp package, which
// applies the name_regex filter.
type regexpValidator struct{}

func (regexpValidator) Description(context.Context) string {
	return "value must be a valid regular expression"
}

func (v regexpValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (regexpValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := regexp.Compile(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid regular expression", err.Error())
	}
}

func validRegexp() validator.String {
	return schemadoc.DescribeString(
		regexpValidator{},
		"The value must be a regular expression in the [RE2 syntax](https://github.com/google/re2/wiki/Syntax).",
	)
}
