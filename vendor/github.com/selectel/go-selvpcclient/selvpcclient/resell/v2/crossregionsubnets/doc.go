/*
Package crossregionsubnets provides the ability to retrieve and manage cross-region subnets
through the Resell v2 API.

Example of getting a single cross-region subnet referenced by its id

  crossRegionSubnet, _, err := crossregionsubnets.Get(context, resellClient, crossRegionSubnetID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(crossRegionSubnet)

Example of getting all cross-region subnets

  allCrossRegionSubnets, _, err := crossregionsubnets.List(ctx, resellClient, crossregionsubnets.ListOpts{})
  if err != nil {
    log.Fatal(err)
  }
  for _, crossRegionSubnet := range allCrossRegionSubnets {
    fmt.Println(crossRegionSubnet)
  }

Example of creating cross-region subnets

  createOpts := crossregionsubnets.CrossRegionSubnetOpts{
    CrossRegionSubnets: []crossregionsubnets.CrossRegionSubnetOpt{
      {
        Quantity: 1,
        Regions: []crossregionsubnets.CrossRegionOpt{
          {
            Region: "ru-1",
          },
          {
            Region: "ru-3",
          },
        },
        CIDR: "192.168.200.0/24",
      },
    },
  }
  newCrossRegionSubnets, _, err := crossregionsubnets.Create(ctx, resellClient, projectID, createOpts)
  if err != nil {
    log.Fatal(err)
  }
  for _, newCrossRegionSubnet := range newCrossRegionSubnets {
    fmt.Printf("%v\n", newCrossRegionSubnet)
  }

Example of deleting a single cross-region subnet

  _, err = crossregionsubnets.Delete(ctx, resellClient, crossRegionSubnetID)
  if err != nil {
    log.Fatal(err)
  }
*/
package crossregionsubnets
