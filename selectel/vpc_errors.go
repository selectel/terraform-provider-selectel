package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
)

func vpcCheckDeleted(d *schema.ResourceData, err error, object string) diag.Diagnostics {
	if err == nil {
		return nil
	}

	if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		d.SetId("")

		return nil
	}

	return diag.FromErr(errGettingObject(object, d.Id(), err))
}
