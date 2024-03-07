package selectel

import (
	"reflect"
	"testing"

	"github.com/selectel/iam-go/service/roles"
)

func TestIAMManageRoles(t *testing.T) {
	type args struct {
		oldRoles []roles.Role
		newRoles []roles.Role
	}
	tests := []struct {
		name           string
		args           args
		wantToUnassign []roles.Role
		wantToAssign   []roles.Role
	}{
		{
			name: "Test assigning new roles",
			args: args{
				oldRoles: []roles.Role{},
				newRoles: []roles.Role{
					{
						RoleName:  "role1",
						Scope:     "scope1",
						ProjectID: "project1",
					},
					{
						RoleName:  "role2",
						Scope:     "scope2",
						ProjectID: "project2",
					},
				},
			},
			wantToUnassign: []roles.Role{},
			wantToAssign: []roles.Role{
				{
					RoleName:  "role1",
					Scope:     "scope1",
					ProjectID: "project1",
				},
				{
					RoleName:  "role2",
					Scope:     "scope2",
					ProjectID: "project2",
				},
			},
		},
		{
			name: "Test unassigning all roles",
			args: args{
				oldRoles: []roles.Role{
					{
						RoleName:  "role1",
						Scope:     "scope1",
						ProjectID: "project1",
					},
					{
						RoleName:  "role2",
						Scope:     "scope2",
						ProjectID: "project2",
					},
				},
				newRoles: []roles.Role{},
			},
			wantToUnassign: []roles.Role{
				{
					RoleName:  "role1",
					Scope:     "scope1",
					ProjectID: "project1",
				},
				{
					RoleName:  "role2",
					Scope:     "scope2",
					ProjectID: "project2",
				},
			},
			wantToAssign: []roles.Role{},
		},
		{
			name: "Test mix of assigning and unassigning roles",
			args: args{
				oldRoles: []roles.Role{
					{
						RoleName:  "role1",
						Scope:     "scope1",
						ProjectID: "project1",
					},
					{
						RoleName:  "role2",
						Scope:     "scope2",
						ProjectID: "project2",
					},
				},
				newRoles: []roles.Role{
					{
						RoleName:  "role2",
						Scope:     "scope2",
						ProjectID: "project2",
					},
					{
						RoleName:  "role3",
						Scope:     "scope3",
						ProjectID: "project3",
					},
					{
						RoleName:  "role4",
						Scope:     "scope4",
						ProjectID: "project4",
					},
				},
			},
			wantToUnassign: []roles.Role{
				{
					RoleName:  "role1",
					Scope:     "scope1",
					ProjectID: "project1",
				},
			},
			wantToAssign: []roles.Role{
				{
					RoleName:  "role3",
					Scope:     "scope3",
					ProjectID: "project3",
				},
				{
					RoleName:  "role4",
					Scope:     "scope4",
					ProjectID: "project4",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualRolesToUnassign, actualRolesToAssign := manageRoles(tt.args.oldRoles, tt.args.newRoles)
			if !reflect.DeepEqual(actualRolesToUnassign, tt.wantToUnassign) {
				t.Errorf("manageRoles() actualRolesToUnassign = %v, want %v", actualRolesToUnassign, tt.wantToUnassign)
			}
			if !reflect.DeepEqual(actualRolesToAssign, tt.wantToAssign) {
				t.Errorf("manageRoles() actualRolesToAssign = %v, want %v", actualRolesToAssign, tt.wantToAssign)
			}
		})
	}
}
