/*
Package users provides the ability to retrieve and manage users through the
Resell v2 API.

Example of getting a single user referenced by its id

  user, _, err := users.Get(context, resellClient, userID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(user)

Example of getting all users

  allUsers, _, err := users.List(ctx, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, user := range allUsers {
    fmt.Println(user)
  }

Example of creating a single user

  userCreateOpts := users.UserOpts{
    Name:     "user0",
    Password: "verysecret",
  }
  createdUser, _, err := users.Create(ctx, resellClient, userCreateOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(createdUser)

Example of updating a single user

  userEnabled := false
  userUpdateOpts := users.UserOpts{
    Name:     "user1",
    Password: "moresecret",
    Enabled:  &userEnabled,
  }
  updatedUser, _, err := users.Update(ctx, resellClient, createdUser.ID, userUpdateOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(updatedUser)

Example of deleting a single user

  _, err = users.Delete(context, resellClient, createdUser.ID)
  if err != nil {
    log.Fatal(err)
  }
*/
package users
