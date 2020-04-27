/*
Package kubeversion provides the ability to retrieve all supported Kubernetes
versions through the MKS V1 API.

Example of getting all supported Kubernetes versions

  kubeVersions, _, err := kubeversion.List(ctx, mksClient)
  if err != nil {
    log.Fatal(err)
  }
  for _, version := range kubeVersions {
    fmt.Printf("%+v\n", version)
  }
*/
package kubeversion
