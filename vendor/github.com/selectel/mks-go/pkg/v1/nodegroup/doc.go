/*
Package nodegroup provides the ability to retrieve and manage cluster nodegroups
through the MKS V1 API.

Example of getting a single cluster nodegroup referenced by its id

  clusterNodegroup, _, err := nodegroup.Get(ctx, mksClient, clusterID, nodegroupID)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%+v\n", clusterNodegroup)

Example of getting all cluster nodegroups

  clusterNodegroups, _, err := nodegroup.List(ctx, mksClient, clusterID)
  if err != nil {
    log.Fatal(err)
  }
  for _, clusterNodegroup := range clusterNodegroups {
    fmt.Printf("%+v\n", clusterNodegroup)
  }

Example of creating a new cluster nodegroup

  createOpts := &nodegroup.CreateOpts{
    Count:            1,
    CPUs:             1,
    RAMMB:            2048,
    VolumeGB:         10,
    VolumeType:       "fast.ru-3a",
    KeypairName:      "ssh-key",
    AvailabilityZone: "ru-3a",
    Labels: map[string]string{
      "label-key0": "label-value0",
      "label-key1": "label-value1",
      "label-key2": "label-value2",
    },
  }
  _, err := nodegroup.Create(ctx, mksClient, clusterID, createOpts)
  if err != nil {
    log.Fatal(err)
  }

Example of deleting a single cluster nodegroup

  _, err := nodegroup.Delete(ctx, mksClient, clusterID, nodegroupID)
  if err != nil {
    log.Fatal(err)
  }

Example of resizing a single cluster nodegroup

  resizeOpts := &nodegroup.ResizeOpts{
    Desired: 1,
  }
  _, err := nodegroup.Resize(ctx, mksClient, clusterID, nodegroupID, resizeOpts)
  if err != nil {
    log.Fatal(err)
  }

Example of updating nodegroup labels

  updateOpts := &nodegroup.UpdateOpts{
    Labels: map[string]string{
      "label-key0": "label-value0",
      "label-key1": "label-value1",
      "label-key2": "label-value2",
    },
  }
  _, err := nodegroup.Update(ctx, mksClient, clusterID, nodegroupID, updateOpts)
  if err != nil {
    log.Fatal(err)
  }
*/
package nodegroup
