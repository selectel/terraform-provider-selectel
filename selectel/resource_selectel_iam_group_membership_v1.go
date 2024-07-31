package selectel

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIAMGroupMembershipV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIAMGroupMembershipV1Create,
		ReadContext:   resourceIAMGroupMembershipV1Read,
		UpdateContext: resourceIAMGroupMembershipV1Update,
		DeleteContext: resourceIAMGroupMembershipV1Delete,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceIAMGroupMembershipV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	userIDsInterface := d.Get("user_ids").([]interface{})
	userIDs := make([]string, len(userIDsInterface))
	for i, v := range userIDsInterface {
		userIDs[i] = v.(string)
	}

	log.Print(msgCreate(objectGroupMembership, d.Id()))
	if len(userIDs) == 0 {
		createErr := fmt.Errorf("error creating group membership: no user ids specified")
		return diag.FromErr(errCreatingObject(objectGroupMembership, createErr))
	}
	err := iamClient.Groups.AddUsers(ctx, d.Get("group_id").(string), userIDs)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectGroupMembership, err))
	}

	d.SetId(d.Get("group_id").(string))

	return resourceIAMGroupMembershipV1Read(ctx, d, meta)
}

func resourceIAMGroupMembershipV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	userIDsInterface := d.Get("user_ids").([]interface{})
	userIDs := make([]string, len(userIDsInterface))
	for i, v := range userIDsInterface {
		userIDs[i] = v.(string)
	}

	response, err := iamClient.Groups.Get(ctx, groupID)
	if err != nil {
		return diag.FromErr(errGettingObject(objectGroupMembership, d.Id(), err))
	}

	responseUserIDs := make([]string, 0)
	for _, user := range response.Users {
		responseUserIDs = append(responseUserIDs, user.KeystoneID)
	}

	responseServiceUserIDs := make([]string, 0)
	for _, serviceUser := range response.ServiceUsers {
		responseServiceUserIDs = append(responseServiceUserIDs, serviceUser.ID)
	}

	d.Set("group_id", groupID)
	d.Set("user_ids", append(responseUserIDs, responseServiceUserIDs...))

	return nil
}

func resourceIAMGroupMembershipV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	oldValue, newValue := d.GetChange("user_ids")

	oldUserIDs := make([]string, len(oldValue.([]interface{})))
	for i, v := range oldValue.([]interface{}) {
		oldUserIDs[i] = v.(string)
	}

	newUserIDs := make([]string, len(newValue.([]interface{})))
	for i, v := range newValue.([]interface{}) {
		newUserIDs[i] = v.(string)
	}

	usersToAdd, usersToRemove := diffUsers(oldUserIDs, newUserIDs)

	if len(usersToAdd) > 0 {
		err := iamClient.Groups.AddUsers(ctx, groupID, usersToAdd)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if len(usersToRemove) > 0 {
		err := iamClient.Groups.DeleteUsers(ctx, groupID, usersToRemove)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(groupID)

	return resourceIAMGroupMembershipV1Read(ctx, d, meta)
}

func resourceIAMGroupMembershipV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamClient, diagErr := getIAMClient(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	userIDsInterface := d.Get("user_ids").([]interface{})
	userIDs := make([]string, len(userIDsInterface))
	for i, v := range userIDsInterface {
		userIDs[i] = v.(string)
	}

	err := iamClient.Groups.DeleteUsers(ctx, groupID, userIDs)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectGroupMembership, d.Id(), err))
	}

	d.SetId("")

	return nil
}

func diffUsers(oldUsers, newUsers []string) ([]string, []string) {
	usersToAdd := make([]string, 0)
	usersToRemove := make([]string, 0)

	for _, user := range newUsers {
		if !slices.Contains(oldUsers, user) {
			usersToAdd = append(usersToAdd, user)
		}
	}

	for _, user := range oldUsers {
		if !slices.Contains(newUsers, user) {
			usersToRemove = append(usersToRemove, user)
		}
	}

	return usersToAdd, usersToRemove
}

// containsAll checks if sliceB is a subset of sliceA.
func containsAll(sliceA, sliceB []string) bool {
	for _, b := range sliceB {
		found := false
		for _, a := range sliceA {
			if a == b {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
