/*
Package node provides the ability to retrieve and manage Kubernetes nodes
of a cluster nodegroup through the MKS V1 API.

Example of getting a single node of a cluster nodegroup by its id

  singleNode, _, err := node.Get(ctx, mksClient, clusterID, nodegroupID, nodeID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", singleNode)

Example of reinstalling a single node of a cluster nodegroup by its id

  _, err := node.Reinstall(ctx, mksClient, clusterID, nodegroupID, nodeID)
  if err != nil {
    log.Fatal(err)
  }

Example of deleting a single node of a cluster nodegroup by its id

  _, err := node.Delete(ctx, mksClient, clusterID, nodegroupID, nodeID)
  if err != nil {
    log.Fatal(err)
  }
*/
package node
