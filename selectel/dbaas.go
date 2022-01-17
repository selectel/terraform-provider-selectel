package selectel

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
)

const (
	ru1DBaaSV1Endpoint = "https://ru-1.dbaas.selcloud.ru/v1"
	ru2DBaaSV1Endpoint = "https://ru-2.dbaas.selcloud.ru/v1"
	ru3DBaaSV1Endpoint = "https://ru-3.dbaas.selcloud.ru/v1"
	ru7DBaaSV1Endpoint = "https://ru-7.dbaas.selcloud.ru/v1"
	ru8DBaaSV1Endpoint = "https://ru-8.dbaas.selcloud.ru/v1"
	ru9DBaaSV1Endpoint = "https://ru-9.dbaas.selcloud.ru/v1"
)

func getDBaaSV1Endpoint(region string) (endpoint string) {
	switch region {
	case ru1Region:
		endpoint = ru1DBaaSV1Endpoint
	case ru2Region:
		endpoint = ru2DBaaSV1Endpoint
	case ru3Region:
		endpoint = ru3DBaaSV1Endpoint
	case ru7Region:
		endpoint = ru7DBaaSV1Endpoint
	case ru8Region:
		endpoint = ru8DBaaSV1Endpoint
	case ru9Region:
		endpoint = ru9DBaaSV1Endpoint
	}

	return
}

func getDBaaSClient(ctx context.Context, d *schema.ResourceData, meta interface{}) (*dbaas.API, diag.Diagnostics) {
	config := meta.(*Config)
	resellV2Client := config.resellV2Client()
	tokenOpts := tokens.TokenOpts{
		ProjectID: d.Get("project_id").(string),
	}

	log.Print(msgCreate(objectToken, tokenOpts))
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return nil, diag.FromErr((errCreatingObject(objectToken, err)))
	}

	region := d.Get("region").(string)
	endpoint := getDBaaSV1Endpoint(region)
	client, err := dbaas.NewDBAASClient(token.ID, endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}

func stringChecksum(s string) (string, error) {
	h := md5.New() // #nosec
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs), nil
}

func stringListChecksum(s []string) (string, error) {
	sort.Strings(s)
	checksum, err := stringChecksum(strings.Join(s, ""))
	if err != nil {
		return "", err
	}

	return checksum, nil
}

func baseTestAccCheckDBaaSV1EntityExists(ctx context.Context, rs *terraform.ResourceState, testAccProvider *schema.Provider) (*dbaas.API, error) {
	var projectID, endpoint string
	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}
	if region, ok := rs.Primary.Attributes["region"]; ok {
		endpoint = getDBaaSV1Endpoint(region)
	}

	config := testAccProvider.Meta().(*Config)
	resellV2Client := config.resellV2Client()

	tokenOpts := tokens.TokenOpts{
		ProjectID: projectID,
	}
	token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
	if err != nil {
		return nil, errCreatingObject(objectToken, err)
	}

	dbaasClient, err := dbaas.NewDBAASClient(token.ID, endpoint)
	if err != nil {
		return nil, err
	}

	return dbaasClient, nil
}

func convertFieldToStringByType(field interface{}) string {
	switch fieldValue := field.(type) {
	case int:
		return strconv.Itoa(fieldValue)
	case float64:
		return strconv.FormatFloat(fieldValue, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(fieldValue), 'f', -1, 32)
	case string:
		return fieldValue
	case bool:
		return strconv.FormatBool(fieldValue)
	default:
		return ""
	}
}

func RandomWithPrefix(name string) string {
	return fmt.Sprintf("%s_%d", name, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
