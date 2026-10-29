package filestorage

import (
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemadoc"
)

// shareLocation is the file storage section of the product availability page.
const shareLocation = "file-storage"

var (
	shareDocs  = schemadoc.Resource{Name: "file storage", Location: shareLocation}
	accessDocs = schemadoc.Resource{Name: "access rule", Location: shareLocation}
)

// The generic driver of the fork boots a service server for every share, so
// creation takes minutes.
var shareTimeouts = schemadoc.Timeouts{
	Create: 30 * time.Minute,
	Update: 20 * time.Minute,
	Delete: 20 * time.Minute,
}

// shareMinSize is the provisioning:min_share_size in GB of every file storage
// type. The maximum differs between pool segments, so the API checks it.
const shareMinSize = 50

const (
	protoNFS  = "NFS"
	protoCIFS = "CIFS"
)

var (
	ipv4Pattern = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func ipv4Validator() validator.String {
	return schemadoc.DescribeString(
		stringvalidator.RegexMatches(ipv4Pattern, "must be an IPv4 address"),
		"The format is an IPv4 address, for example, `192.168.199.10`.",
	)
}

const createTokenKey = "selectel_tf_create_token" //nolint:gosec // Metadata key name, not a credential.

// reservedMetadataKeys are set by the fork or the provider. A configured value
// would overwrite or remove what the service relies on, such as the flag that
// deletes the service network together with the share.
var reservedMetadataKeys = []string{
	"delete_empty_share_network",
	"neutron_subnet_id",
	"volume_type",
	"access_list_allow_all",
	"ip",
	createTokenKey,
}
