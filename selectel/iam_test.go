package selectel

import (
	"testing"

	"github.com/selectel/iam-go/service/roles"
	"github.com/stretchr/testify/assert"
)

func TestIAMDiffRoles(t *testing.T) {
	type args struct {
		oldRoles []roles.Role
		newRoles []roles.Role
	}
	tests := map[string]struct {
		args           args
		wantToUnassign []roles.Role
		wantToAssign   []roles.Role
	}{
		"Test assigning new roles": {
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
		"Test unassigning all roles": {
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
		"Test mix of assigning and unassigning roles": {
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

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualRolesToUnassign, actualRolesToAssign := diffRoles(tt.args.oldRoles, tt.args.newRoles)
			assert := assert.New(t)
			assert.Equal(tt.wantToUnassign, actualRolesToUnassign)
			assert.Equal(tt.wantToAssign, actualRolesToAssign)
		})
	}
}
