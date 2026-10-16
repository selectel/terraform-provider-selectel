package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestMain runs the tests, or only the sweepers when -sweep is set.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
