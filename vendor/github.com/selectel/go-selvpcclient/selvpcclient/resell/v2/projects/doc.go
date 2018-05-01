/*
Package projects provides the ability to retrieve and manage projects through the
Resell v2 API.

Example of getting a single project referenced by its id

  project, _, err := projects.Get(context, resellClient, projectID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(project)

Example of listing all projects in the domain

  allProjects, _, err := projects.List(context, resellClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, myProject := range allProjects {
    fmt.Println(myProject)
  }

Example of creating a single project

  createOpts := projects.CreateOpts{
    Name: "test000",
    Quotas: []quotas.QuotaOpts{
      {
        Name: "compute_cores",
        ResourceQuotasOpts: []quotas.ResourceQuotaOpts{
          {
            Region: "ru-1",
            Zone:   "ru-1b",
            Value:  1,
          },
        },
      },
      {
        Name: "compute_ram",
        ResourceQuotasOpts: []quotas.ResourceQuotaOpts{
          {
            Region: "ru-1",
            Zone:   "ru-1b",
            Value:  1024,
          },
        },
      },
    },
  }
  newProject, _, err := projects.Create(context, resellClient, createOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(newProject)

Example of updating a single project

  themeColor := "ffffff"
  logo := "123"
  themeUpdateOpts := projects.ThemeUpdateOpts{
    Color: &themeColor,
    Logo:  &logo,
  }
  name := "test001"
  updateOpts := projects.UpdateOpts{
    Name:  &name,
    Theme: &themeUpdateOpts,
  }
  updatedProject, _, err := projects.Update(context, resellClient, newProject.ID, updateOpts)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(updatedProject)

Example of deleting a single project

  _, err = projects.Delete(context, resellClient, newProject.ID)
  if err != nil {
    log.Fatal(err)
  }
*/
package projects
