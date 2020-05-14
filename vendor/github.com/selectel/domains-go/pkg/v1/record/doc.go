/*
Package record provides the ability to interact with domain records
through the Selectel Domains API V1.

Example of getting a single domain record by its id

  domainRecord, _, err := record.Get(ctx, serviceClient, domainID, recordID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", domainRecord)

Example of getting a list of domain records by domain id

  domainRecords, _, err := record.ListByDomainID(ctx, serviceClient, domainID)
  if err != nil {
    log.Fatal(err)
  }
  for _, r := range domainRecords {
    fmt.Printf("%+v\n", r)
  }

Example of getting a list of domain records by domain name

  domainRecords, _, err := record.ListByDomainName(ctx, serviceClient, domainName)
  if err != nil {
    log.Fatal(err)
  }
  for _, r := range domainRecords {
    fmt.Printf("%+v\n", r)
  }

Example of creating a new domain record

  createOpts := &record.CreateOpts{
    Name:     "example.testdomain.xyz",
    Type:     record.TypeCNAME,
    TTL:      60,
    Content:  "origin.example.com",
  }
  domainRecord, _, err := record.Create(ctx, serviceClient, domainID, createOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", domainRecord)

Example of domain record deletion

  _, err := record.Delete(ctx, serviceClient, domainID, recordID)
  if err != nil {
    log.Fatal(err)
  }

Example of domain record updating

  updateOpts := &record.UpdateOpts{
    Name:     "example.testdomain.xyz",
    Type:     record.TypeCNAME,
    TTL:      120,
    Content:  "origin.example.com",
  }
  updatedRecord, _, err := record.Update(ctx, serviceClient, domainID, recordID, updateOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", updatedRecord)
*/
package record
