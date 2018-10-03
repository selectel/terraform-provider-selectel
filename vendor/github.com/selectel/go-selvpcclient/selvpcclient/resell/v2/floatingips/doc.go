/*
Package floatingips provides the ability to retrieve and manage floating ips through
the Resell v2 API.

Example of getting a single floating ip referenced by its id

  floatingIP, _, err := floatingips.Get(context, resellClient, fipID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(floatingIP)

Example of getting all floating ips

  allFloatingIPs, _, err := floatingips.List(ctx, resellClient, floatingips.ListOpts{})
  if err != nil {
    log.Fatal(err)
  }
  for _, floatingIP := range floatingips {
    fmt.Println(floatingIP)
  }

Example of creating floating ips in a project

  newFloatingIPsOpts := floatingips.FloatingIPOpts{
    FloatingIPs: []floatingips.FloatingIPOpt{
      {
        Region:   "ru-2",
        Quantity: 2,
      },
    },
  }
  projectID := "49338ac045f448e294b25d013f890317"
  newFloatingIPs, _, err := floatingips.Create(ctx, resellClient, projectID, newFloatingIPsOpts)
  if err != nil {
    log.Fatal(err)
  }
  for _, newFloatingIP := range newFloatingIPs {
    fmt.Println(newFloatingIPs)
  }

Example of deleting a single floating ip

  _, err = floatingips.Delete(ctx, resellClient, "412a04ba-4cb2-4823-abd1-fcd48952b882")
  if err != nil {
    log.Fatal(err)
  }
*/
package floatingips
