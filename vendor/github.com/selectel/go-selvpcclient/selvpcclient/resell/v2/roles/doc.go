/*
Package roles provides the ability to retrieve and manage roles through the
Resell v2 API.

Example of getting roles in the specified project

  allRoles, _, err := roles.ListProject(context, resellClient, projectID)
  if err != nil {
    log.Fatal(err)
  }
  for _, myRole := range allRoles {
    fmt.Println(myRole)
  }

Example of getting roles for the specified user

  allRoles, _, err := roles.ListUser(context, resellClient, userID)
  if err != nil {
    log.Fatal(err)
  }
  for _, myRole := range allRoles {
    fmt.Println(myRole)
  }

Example of creating a single role

  createOpts := roles.RoleOpt{
    ProjectID: "49338ac045f448e294b25d013f890317",
    UserID:    "763eecfaeb0c8e9b76ab12a82eb4c11",
  }
  role, _, err := roles.Create(ctx, resellClient, createOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(myRole)

Example of creating several roles

  createOpts := roles.RoleOpts{
    Roles: []roles.RoleOpt{
      {
        ProjectID: "81800a8ec3fc49fca2cf00857de3ae9d",
        UserID:    "763eecfaeb0c8e9b76ab12a82eb4c11",
      },
      {
        ProjectID: "d7452adc9769422a908edfd2281d7c55",
        UserID:    "763eecfaeb0c8e9b76ab12a82eb4c11",
      },
    },
  }
  allRoles, httpResponse, err := roles.CreateBulk(ctx, testEnv.Client, createOpts)
  for _, myRole := range allRoles {
    fmt.Println(myRole)
  }


Example of deleting a single role

  deleteOpts := roles.RoleOpt{
    ProjectID: "49338ac045f448e294b25d013f890317",
    UserID:    "763eecfaeb0c8e9b76ab12a82eb4c11",
  }
  _, err := roles.Delete(ctx, resellClient, deleteOpts)
  if err != nil {
    log.Fatal(err)
  }
*/
package roles
