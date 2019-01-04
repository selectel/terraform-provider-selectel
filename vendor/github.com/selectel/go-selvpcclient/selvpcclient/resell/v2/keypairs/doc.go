/*
Package keypairs provides the ability to retrieve and manage keypairs through
the Resell v2 API.

Example of getting keypairs in the current domain

  allKeypairs, _, err = keypairs.List(context, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, myKeypair := range allKeypairs {
    fmt.Println(myKeypair)
  }

Example of creating keypairs in all regions with the same options

  newKeypairOptions := keypairs.KeypairOpts{
    Name:      "my_keypair",
    PublicKey: "ssh-rsa public_key_part user0@example.org",
    UserID:    "82a026cae2104e92b999dbe00cdb9435",
  }
  newKeypairs, _, err := keypairs.Create(ctx, resellClient, newKeypairOptions)
  if err != nil {
    log.Fatal(err)
  }
  for _, newKeypair := range newKeypairs {
    fmt.Printf("%v\n", newKeypair)
  }

Example of deleting a single keypair of a user

  keypairName := "my_keypair"
  userID := 82a026cae2104e92b999dbe00cdb9435""
  _, err = keypairs.Delete(ctx, resellClient, keypairName, userID)
  if err != nil {
    log.Fatal(err)
  }
*/
package keypairs
