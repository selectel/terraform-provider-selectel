package images

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/selectel/images-go/pkg/v1/member"
)

// membership is the part of the image access models that the API reports.
// The owner of the image and the project that received access manage the same
// membership with different resources.
type membership struct {
	ID        types.String `tfsdk:"id"`
	ImageID   types.String `tfsdk:"image_id"`
	MemberID  types.String `tfsdk:"member_id"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (m *membership) observe(current *member.View) {
	m.ID = types.StringValue(membershipID(current.ImageID, current.MemberID))
	m.ImageID = types.StringValue(current.ImageID)
	m.MemberID = types.StringValue(current.MemberID)
	m.Status = types.StringValue(current.Status)
	m.CreatedAt = types.StringValue(current.CreatedAt.Format(time.RFC3339))
	m.UpdatedAt = types.StringValue(current.UpdatedAt.Format(time.RFC3339))
}

// membershipID is the ID of both access resources: a membership has no ID of
// its own and is addressed by the image and the project that received access.
func membershipID(imageID, memberID string) string { return imageID + "/" + memberID }

// parseMembershipID splits an import ID. Without memberRequired, the ID may
// be the image ID alone, and the member ID is empty.
func parseMembershipID(id string, memberRequired bool) (string, string, error) {
	imageID, memberID, hasMember := strings.Cut(id, "/")
	switch {
	case !isImageID(imageID):
		return "", "", fmt.Errorf("the image ID %q is not a lowercase UUID", imageID)
	case hasMember && (memberID == "" || strings.Contains(memberID, "/")),
		!hasMember && memberRequired:
		return "", "", fmt.Errorf("the ID %q must have the <image_id>/<member_id> format", id)
	}

	return imageID, memberID, nil
}
