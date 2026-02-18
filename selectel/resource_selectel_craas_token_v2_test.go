package selectel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/craas-go/pkg/svc"
	tokenv2 "github.com/selectel/craas-go/pkg/v2/token"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
)

func TestAccCRaaSTokenV2Basic(t *testing.T) {
	var (
		project    projects.Project
		craasToken tokenv2.TokenV2
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	tokenName := acctest.RandomWithPrefix("tf-acc-token")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCRaaSTokenV2Basic(projectName, tokenName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckCRaaSTokenV2Exists("selectel_craas_token_v2.token_tf_acc_test_1", &craasToken),
					resource.TestCheckResourceAttr("selectel_craas_token_v2.token_tf_acc_test_1", "name", tokenName),
					resource.TestCheckResourceAttr("selectel_craas_token_v2.token_tf_acc_test_1", "username", craasV1TokenUsername),
					resource.TestCheckResourceAttrSet("selectel_craas_token_v2.token_tf_acc_test_1", "token"),
				),
			},
			{
				Config: testAccCRaaSTokenV2Update(projectName, tokenName+"-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckCRaaSTokenV2Exists("selectel_craas_token_v2.token_tf_acc_test_1", &craasToken),
					resource.TestCheckResourceAttr("selectel_craas_token_v2.token_tf_acc_test_1", "name", tokenName+"-updated"),
				),
			},
		},
	})
}

func testAccCheckCRaaSTokenV2Exists(n string, craasToken *tokenv2.TokenV2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		ctx := context.Background()
		craasClient, err := newCRaaSV2TestClient(rs, testAccProvider)
		if err != nil {
			return err
		}

		foundToken, _, err := tokenv2.GetByID(ctx, craasClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundToken.ID != rs.Primary.ID {
			return errors.New("token not found")
		}

		*craasToken = *foundToken

		return nil
	}
}

func testAccCRaaSTokenV2Basic(projectName, tokenName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_craas_token_v2" "token_tf_acc_test_1" {
  project_id     = selectel_vpc_project_v2.project_tf_acc_test_1.id
  name           = "%s"
  mode_rw        = true
  all_registries = true
  is_set         = true
  expires_at     = "2030-01-01T00:00:00Z"
}
`, projectName, tokenName)
}

func testAccCRaaSTokenV2Update(projectName, tokenName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name = "%s"
}

resource "selectel_craas_token_v2" "token_tf_acc_test_1" {
  project_id     = selectel_vpc_project_v2.project_tf_acc_test_1.id
  name           = "%s"
  mode_rw        = true
  all_registries = true
  is_set         = true
  expires_at     = "2030-01-01T00:00:00Z"
}
`, projectName, tokenName)
}

func TestShouldRemoveCRaaSTokenV2FromStateAt(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	futureExpiry := now.Add(365 * 24 * time.Hour)
	farFutureExpiry := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	pastExpiry := now.Add(-24 * time.Hour)
	farPastExpiry := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	justExpired := now.Add(-1 * time.Second)

	tests := []struct {
		name         string
		token        *tokenv2.TokenV2
		wantRemove   bool
		wantReasonRe string
	}{
		{
			name: "active token with future expiry stays in state",
			token: &tokenv2.TokenV2{
				ID:     "tok-1",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: futureExpiry,
				},
			},
			wantRemove: false,
		},
		{
			name: "active token with far-future expiry stays in state",
			token: &tokenv2.TokenV2{
				ID:     "tok-1b",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: farFutureExpiry,
				},
			},
			wantRemove: false,
		},
		{
			name: "active token without expiry stays in state",
			token: &tokenv2.TokenV2{
				ID:     "tok-2",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet: false,
				},
			},
			wantRemove: false,
		},
		{
			name: "active token with zero expiry and isSet false stays in state",
			token: &tokenv2.TokenV2{
				ID:     "tok-2b",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     false,
					ExpiresAt: time.Time{},
				},
			},
			wantRemove: false,
		},
		{
			name: "active token with expiry equal to now is not removed",
			token: &tokenv2.TokenV2{
				ID:     "tok-2c",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: now,
				},
			},
			wantRemove: false,
		},
		{
			name: "revoked token removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-3",
				Status: "revoked",
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: futureExpiry,
				},
			},
			wantRemove:   true,
			wantReasonRe: `non-active status "revoked"`,
		},
		{
			name: "revoked token with far-future expiry removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-3b",
				Status: "revoked",
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: farFutureExpiry,
				},
			},
			wantRemove:   true,
			wantReasonRe: `non-active status "revoked"`,
		},
		{
			name: "expired status token removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-4",
				Status: tokenv2.StatusExpired,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: pastExpiry,
				},
			},
			wantRemove:   true,
			wantReasonRe: `non-active status "expired"`,
		},
		{
			name: "deleted status token removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-5",
				Status: tokenv2.StatusDeleted,
				Expiration: tokenv2.Expiration{
					IsSet: false,
				},
			},
			wantRemove:   true,
			wantReasonRe: `non-active status "deleted"`,
		},
		{
			name: "active token with past expiry removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-6",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: pastExpiry,
				},
			},
			wantRemove:   true,
			wantReasonRe: "expired at",
		},
		{
			name: "active token with far-past expiry removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-6b",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: farPastExpiry,
				},
			},
			wantRemove:   true,
			wantReasonRe: "expired at",
		},
		{
			name: "active token that just expired removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-7",
				Status: tokenv2.StatusActive,
				Expiration: tokenv2.Expiration{
					IsSet:     true,
					ExpiresAt: justExpired,
				},
			},
			wantRemove:   true,
			wantReasonRe: "expired at",
		},
		{
			name: "unknown status token removed from state",
			token: &tokenv2.TokenV2{
				ID:     "tok-8",
				Status: "some-unknown-status",
				Expiration: tokenv2.Expiration{
					IsSet: false,
				},
			},
			wantRemove:   true,
			wantReasonRe: `non-active status "some-unknown-status"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemove, gotReason := shouldRemoveCRaaSTokenV2FromStateAt(tt.token, now)

			assert.Equal(t, tt.wantRemove, gotRemove, "removal decision mismatch")

			if tt.wantRemove {
				assert.Contains(t, gotReason, tt.wantReasonRe,
					"reason should contain expected substring")
			} else {
				assert.Empty(t, gotReason,
					"reason should be empty when token stays in state")
			}
		})
	}
}

func TestIsCRaaSTokenV2DeleteNotFound(t *testing.T) {
	tests := []struct {
		name     string
		response *svc.ResponseResult
		want     bool
	}{
		{
			name:     "nil response is not a 404",
			response: nil,
			want:     false,
		},
		{
			name: "404 response is treated as already deleted",
			response: &svc.ResponseResult{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
			want: true,
		},
		{
			name: "500 response is not a 404",
			response: &svc.ResponseResult{
				Response: &http.Response{
					StatusCode: http.StatusInternalServerError,
				},
			},
			want: false,
		},
		{
			name: "204 response is not a 404",
			response: &svc.ResponseResult{
				Response: &http.Response{
					StatusCode: http.StatusNoContent,
				},
			},
			want: false,
		},
		{
			name: "403 response is not a 404",
			response: &svc.ResponseResult{
				Response: &http.Response{
					StatusCode: http.StatusForbidden,
				},
			},
			want: false,
		},
		{
			name: "200 response is not a 404",
			response: &svc.ResponseResult{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCRaaSTokenV2DeleteNotFound(tt.response)
			assert.Equal(t, tt.want, got)
		})
	}
}
