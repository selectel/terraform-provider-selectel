/*
Package domain provides the ability to interact with domains
through the Selectel Domains API V1.

Example of getting a single domain by its id

  selectelDomain, _, err := domain.GetByID(ctx, serviceClient, domainID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", selectelDomain)

Example of getting a single domain by its name

  selectelDomain, _, err := domain.GetByName(ctx, serviceClient, domainName)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", selectelDomain)

Example of getting a list of domains

  selectelDomains, _, err := domain.List(ctx, serviceClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, d := range selectelDomains {
    fmt.Printf("%+v\n", d)
  }

Example of creating a new domain

  createOpts := &domain.CreateOpts{
    Name:        "testdomain.xyz",
  }
  selectelDomain, _, err := domain.Create(ctx, serviceClient, createOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", selectelDomain)

Example of domain deletion

  _, err := domain.Delete(ctx, serviceClient, domainID)
  if err != nil {
    log.Fatal(err)
  }
*/
package domain
