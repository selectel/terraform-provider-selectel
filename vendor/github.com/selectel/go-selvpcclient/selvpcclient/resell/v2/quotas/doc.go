/*
Package quotas provides the ability to retrieve and update quotas through the
Resell v2 API.

Example of getting all quotas for a domain

  allQuotas, _, err := quotas.GetAll(ctx, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, myQuota := range allQuotas {
    fmt.Println(myQuota)
  }

Example of getting free quotas for a domain

  freeQuotas, _, err := quotas.GetFree(ctx, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, myQuota := range allQuotas {
    fmt.Println(myQuota)
  }

Example of getting projects quotas for a domain

  projectsQuotas, _, err := quotas.GetProjectsQuotas(ctx, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, projectQuota := range projectsQuotas {
    fmt.Printf("quotas for %s:\n", projectQuota.ID)
    for _, resourceQuota := range projectQuota.ProjectQuotas {
      fmt.Println(resourceQuota)
    }
  }

Example of getting quotas for a single project

  singleProjectQuotas, _, err := quotas.GetProjectQuotas(ctx, resellClient, updateProjectID)
  if err != nil {
    log.Fatal(err)
  }
  for _, singleProjectQuota := range singleProjectQuotas {
    fmt.Println(singleProjectQuota)
  }

Example of updating quotas for a single project

  projectQuotaUpdateOpts := quotas.UpdateProjectQuotasOpts{
    QuotasOpts: []*quotas.QuotaOpts{
      {
        Name: "image_gigabytes",
        ResourceQuotasOpts: []quotas.ResourceQuotaOpts{
          {
            Region: "ru-1",
            Value:  10,
          },
          {
            Region: "ru-2",
            Value:  20,
          },
        },
      },
    },
  }
  updatedProjectQuotas, _, err := quotas.UpdateProjectQuotas(context, resellClient, updateProjectID, projectQuotaUpdateOpts)
  if err != nil {
    log.Fatal(err)
  }
  for _, updatedProjectQuota := range updatedProjectQuotas {
    fmt.Println(updatedProjectQuota)
  }
*/
package quotas
