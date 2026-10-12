package blockstorage

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

var volumeTimeouts = schemadoc.Timeouts{
	Create: 10 * time.Minute,
	Update: 20 * time.Minute,
	Delete: 10 * time.Minute,
}

// volumeLocation is the network volumes section of the product availability page.
const volumeLocation = "network-volumes"

func volumeTypeIDValidator() validator.String {
	return schemadoc.DescribeString(stringvalidator.Any(
		schemadoc.UUID(),
		schemadoc.OneOf("default"),
	), "The format is a lowercase UUID or `default` to select the default volume type.")
}
