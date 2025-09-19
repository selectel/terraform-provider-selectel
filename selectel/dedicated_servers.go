package selectel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used
	"net"
	"slices"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/servers"
	serverslocal "github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicatedservers"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/hashcode"
)

// serversMapsFromStructs converts the provided license.Servers to
// the slice of maps correspondingly to the resource's schema.
func serversMapsFromStructs(serverStructs []servers.Server) []map[string]interface{} {
	associatedServers := make([]map[string]interface{}, len(serverStructs))

	for i, server := range serverStructs {
		associatedServers[i] = map[string]interface{}{
			"id":     server.ID,
			"name":   server.Name,
			"status": server.Status,
		}
	}

	return associatedServers
}

// hashServers is a hash function to use with the "servers" set.
func hashServers(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-", m["id"].(string)))
}

func getDedicatedServersClient(d *schema.ResourceData, meta interface{}) (*serverslocal.ServiceClient, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get(dedicatedServersServerSchemaKeyProjectID).(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for dedicated servers api: %w", err))
	}

	url := "https://api.selectel.ru/servers/v2"

	return serverslocal.NewClientV2(selvpcClient.GetXAuthToken(), url), nil
}

type (
	PartitionsConfig struct {
		SoftRaidConfig []*SoftRaidConfigItem
		DiskPartitions []*DiskPartitionsItem
	}

	SoftRaidConfigItem struct {
		Name     string
		Level    string
		DiskType string
	}

	DiskPartitionsItem struct {
		Mount       string
		Size        float64
		SizePercent float64
		Raid        string
		FSType      string
	}
)

func (pc *PartitionsConfig) IsEmpty() bool {
	return len(pc.SoftRaidConfig) == 0 && len(pc.DiskPartitions) == 0
}

const (
	mountBaseBoot = "/boot"
	mountBaseRoot = "/"
	mountBaseSwap = "swap"
)

func (pc *PartitionsConfig) CastToAPIPartitionsConfig(
	localDrives serverslocal.LocalDrives, defaultPartitions []*serverslocal.PartitionConfigItem,
) (serverslocal.PartitionsConfig, error) {
	// if empty and partitioning is on - filling with default values
	if pc.IsEmpty() {
		if len(localDrives) == 0 {
			return nil, errors.New("local drives are required for automatic partitioning")
		}

		raidName := "first-raid"

		pc.SoftRaidConfig = append(pc.SoftRaidConfig, &SoftRaidConfigItem{
			Name:     raidName,
			Level:    "raid1",
			DiskType: localDrives.GetDefaultType(),
		})

		for _, p := range defaultPartitions {
			size := p.Size
			if p.Mount == "/" {
				size = -1
			}

			pc.DiskPartitions = append(pc.DiskPartitions, &DiskPartitionsItem{
				Mount:  p.Mount,
				Size:   size,
				Raid:   raidName,
				FSType: p.FSType,
			})
		}
	}

	// starting to build api config

	res := make(serverslocal.PartitionsConfig)

	// adding local drives
	for ldID, ld := range localDrives {
		res[ldID] = &serverslocal.PartitionConfigItem{
			Type: ld.Type,
			Match: &serverslocal.PartitionConfigItemMatch{
				Size: ld.Match.Size,
				Type: ld.Match.Type,
			},
		}
	}

	// adding boot partition if not presented

	if !pc.ContainsBootPartition() {
		found := false
		for _, dp := range defaultPartitions {
			if dp.Mount == mountBaseBoot {
				pc.DiskPartitions = append(pc.DiskPartitions, &DiskPartitionsItem{
					Raid:   pc.PickDefaultBootRaidName(localDrives),
					Mount:  dp.Mount,
					Size:   dp.Size,
					FSType: dp.FSType,
				})

				found = true

				break
			}
		}

		if !found {
			return nil, errors.New("can't find default partition for boot partition")
		}
	}

	var (
		diskPartitions = make([]*DiskPartitionsItem, 0, len(pc.DiskPartitions))

		nextPriorityByDrive = make(map[string]int)
	)

	// adding non base partitions, soft raids and filesystems

	for _, diskPartition := range pc.DiskPartitions {
		isBasePartition := slices.Contains([]string{
			mountBaseSwap, mountBaseRoot, mountBaseBoot,
		}, diskPartition.Mount)

		if !isBasePartition {
			err := pc.addDiskPartitionToAPIConfig(diskPartition, localDrives, nextPriorityByDrive, res)
			if err != nil {
				return nil, err
			}

			continue
		}

		diskPartitions = append(diskPartitions, diskPartition)
	}

	// adding base partitions, soft raids and filesystems

	for _, diskPartition := range diskPartitions {
		err := pc.addDiskPartitionToAPIConfig(diskPartition, localDrives, nextPriorityByDrive, res)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (pc *PartitionsConfig) addDiskPartitionToAPIConfig(
	diskPartition *DiskPartitionsItem, localDrives serverslocal.LocalDrives, nextPriorityByDrive map[string]int,
	cfg serverslocal.PartitionsConfig,
) error {
	var diskType, raidLevel string

	for _, srCfg := range pc.SoftRaidConfig {
		if srCfg.Name == diskPartition.Raid {
			diskType = srCfg.DiskType
			raidLevel = srCfg.Level

			break
		}
	}

	if diskType == "" {
		return errors.New("can't find disk type for " + diskPartition.Raid)
	}

	// adding partitions

	members := make([]string, 0)

	for ldID, ld := range localDrives {
		if ld.Match.Type != diskType {
			continue
		}

		id, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate uuid for partition %s local drive %s: %w", diskPartition.Mount, ldID, err)
		}

		size := diskPartition.Size
		if diskPartition.SizePercent > 0 {
			baseSize := float64(ld.Match.Size)
			if baseSize <= 0 {
				return fmt.Errorf("invalid local drive %s size: %d", ld.Match.Type, ld.Match.Size)
			}

			percent := diskPartition.SizePercent

			size = math.Round(baseSize * percent / 100.0)
		}

		priority := nextPriorityByDrive[ldID]

		cfg[id] = &serverslocal.PartitionConfigItem{
			Type:     "partition",
			Device:   ldID,
			Size:     size,
			Priority: &priority,
		}

		nextPriorityByDrive[ldID] = priority + 1

		members = append(members, id)
	}

	// adding soft raid

	fsPartitionDeviceID := ""

	switch {
	case len(members) == 0:
		return errors.New("can't find disk for " + diskPartition.Mount)

	case len(members) == 1:
		fsPartitionDeviceID = members[0]

	default:
		srID, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate uuid for soft_raid %s: %w", diskPartition.Mount, err)
		}

		cfg[srID] = &serverslocal.PartitionConfigItem{
			Type:    "soft_raid",
			Members: members,
			Level:   raidLevel,
		}

		fsPartitionDeviceID = srID
	}

	// adding filesystem

	id, err := uuid.GenerateUUID()
	if err != nil {
		return fmt.Errorf("failed to generate uuid for filesystem partition %s: %w", diskPartition.Mount, err)
	}

	fsType := "ext4"

	switch {
	case diskPartition.Mount == mountBaseBoot:
		fsType = "ext3"

	case diskPartition.Mount == mountBaseSwap:
		fsType = "swap"

	case diskPartition.FSType != "":
		fsType = diskPartition.FSType
	}

	cfg[id] = &serverslocal.PartitionConfigItem{
		Type:   "filesystem",
		FSType: fsType,
		Device: fsPartitionDeviceID,
		Mount:  diskPartition.Mount,
	}

	return nil
}

func (pc *PartitionsConfig) ContainsBootPartition() bool {
	for _, dp := range pc.DiskPartitions {
		if dp.Mount == mountBaseBoot {
			return true
		}
	}

	return false
}

func (pc *PartitionsConfig) PickDefaultBootRaidName(localDrives serverslocal.LocalDrives) string {
	var (
		fastestSR      = ""
		fastestSRRatio = 0
	)
	for _, sr := range pc.SoftRaidConfig {
		currRatio := 0
		for _, ld := range localDrives {
			if ld.Match.Type != sr.DiskType {
				continue
			}

			currRatio = ld.SpeedRatio()

			break
		}

		switch {
		case fastestSR == "":
			fastestSR = sr.Name
			fastestSRRatio = currRatio

		case currRatio > fastestSRRatio:
			fastestSR = sr.Name
			fastestSRRatio = currRatio
		}
	}

	return fastestSR
}

func resourceDedicatedServersServerV1ReadPartitionsConfig(d *schema.ResourceData) (*PartitionsConfig, error) {
	res := new(PartitionsConfig)

	v, ok := d.GetOk(dedicatedServersServerSchemaKeyOSPartitionsConfig)
	if !ok {
		return res, nil
	}

	partitionsConfigRaw, ok := v.([]interface{})
	if !ok || len(partitionsConfigRaw) != 1 {
		return nil, errors.New("partitions_config has unexpected type")
	}

	partitionsConfig, ok := partitionsConfigRaw[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("partitions_config has unexpected type")
	}

	var err error
	res.SoftRaidConfig, err = resourceDedicatedServersServerV1ReadPartitionsConfigSoftRaid(partitionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read soft raid configuration: %w", err)
	}

	res.DiskPartitions, err = resourceDedicatedServersServerV1ReadPartitionsConfigDiskPartitions(partitionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read disk partitions configuration: %w", err)
	}

	return res, nil
}

func resourceDedicatedServersServerV1ReadPartitionsConfigSoftRaid(partitionsConfig map[string]interface{}) ([]*SoftRaidConfigItem, error) {
	srCfgRaw, ok := partitionsConfig[dedicatedServersServerSchemaKeySoftRaidConfig].([]interface{})
	if !ok {
		srCfgRaw = make([]interface{}, 0)
	}

	res := make([]*SoftRaidConfigItem, 0, len(srCfgRaw))

	for idx, itemRaw := range srCfgRaw {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d] has unexpected type", idx)
		}

		name, ok := item[dedicatedServersServerSchemaKeyName].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].name has unexpected type", idx)
		}

		level, ok := item[dedicatedServersServerSchemaKeyLevel].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].level has unexpected type", idx)
		}

		diskType, ok := item[dedicatedServersServerSchemaKeyDiskType].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].disk_type has unexpected type", idx)
		}

		res = append(res, &SoftRaidConfigItem{
			Name:     name,
			Level:    level,
			DiskType: diskType,
		})
	}

	return res, nil
}

func resourceDedicatedServersServerV1ReadPartitionsConfigDiskPartitions(partitionsConfig map[string]interface{}) ([]*DiskPartitionsItem, error) {
	dpCfgRaw, ok := partitionsConfig[dedicatedServersServerSchemaKeyDiskPartitions].([]interface{})
	if !ok {
		dpCfgRaw = make([]interface{}, 0)
	}

	res := make([]*DiskPartitionsItem, 0, len(dpCfgRaw))

	for idx, itemRaw := range dpCfgRaw {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d] has unexpected type", idx)
		}

		mount, ok := item[dedicatedServersServerSchemaKeyMount].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d].mount has unexpected type", idx)
		}

		size := 0.0

		sizeRaw, hasSizeRaw := item[dedicatedServersServerSchemaKeySize]
		if hasSizeRaw {
			size, ok = sizeRaw.(float64)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].size has unexpected type", idx)
			}
		}

		sizePercent := 0.0

		sizePercentRaw, hasSizePercentRaw := item[dedicatedServersServerSchemaKeySizePercent]
		if hasSizePercentRaw {
			sizePercent, ok = sizePercentRaw.(float64)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].size_percent has unexpected type", idx)
			}
		}

		if (size != 0 && sizePercent != 0) || (size == 0 && sizePercent == 0) {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d]: size or size_percent must be presented once", idx)
		}

		raid, ok := item[dedicatedServersServerSchemaKeyRaid].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d].raid has unexpected type", idx)
		}

		fsTypeRaw, ok := item[dedicatedServersServerSchemaKeyFSType]

		fsType := ""
		if ok {
			fsType, ok = fsTypeRaw.(string)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].fsType has unexpected type", idx)
			}
		}

		res = append(res, &DiskPartitionsItem{
			Mount:       mount,
			Size:        size,
			SizePercent: sizePercent,
			Raid:        raid,
			FSType:      fsType,
		})
	}

	return res, nil
}

func resourceDedicatedServersServerV1GetFreePublicIPs(
	ctx context.Context, cl *serverslocal.ServiceClient, locationID, subnetID string,
) (net.IP, error) {
	subnet, _, err := cl.NetworkSubnet(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get subnets for %s %s: %w", objectLocation, locationID, err,
		)
	}

	nets, _, err := cl.Networks(ctx, locationID, serverslocal.NetworkTypeInet)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s networks for %s %s: %w", serverslocal.NetworkTypeInet, objectLocation, locationID, err,
		)
	}

	nets = nets.FilterByTelematicsTypeHosting()

	netsContainSubnet := slices.ContainsFunc(nets, func(n *serverslocal.Network) bool {
		return subnet.NetworkUUID == n.UUID
	})

	if !netsContainSubnet || subnet.Free == 0 {
		return nil, fmt.Errorf(
			"subnet %s is not suitable for allocating ip", subnetID,
		)
	}

	reservedIPs, _, err := cl.NetworkReservedIPs(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get reserved ips for %s %s: %w", objectLocation, locationID, err,
		)
	}

	freeIP, err := subnet.GetFreeIP(reservedIPs, false)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to compute free ips for %s %s: %w", objectLocation, locationID, err,
		)
	}

	return freeIP, nil
}

func resourceDedicatedServersServerV1GetFreePrivateIPs(
	ctx context.Context, cl *serverslocal.ServiceClient, locationID, subnetStr string,
) (ip net.IP, subnetID string, err error) {
	nets, _, err := cl.Networks(ctx, locationID, serverslocal.NetworkTypeLocal)
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to get %s networks for %s %s: %w", serverslocal.NetworkTypeInet, objectLocation, locationID, err,
		)
	}

	if len(nets) != 1 {
		return nil, "", fmt.Errorf(
			"expected exactly one local network for %s %s, got %d", objectLocation, locationID, len(nets),
		)
	}

	subnets, _, err := cl.NetworkLocalSubnets(ctx, nets[0].UUID)
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to get subnets for %s %s: %w", objectLocation, locationID, err,
		)
	}

	subnet := subnets.FindBySubnet(subnetStr)
	if subnet == nil {
		return nil, "", fmt.Errorf(
			"can't find subnet %s for %s %s and network %s", subnetStr, objectLocation, locationID, nets[0].UUID,
		)
	}

	reservedIPs, _, err := cl.NetworkSubnetLocalReservedIPs(ctx, subnet.UUID)
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to get reserved ips for %s %s: %w", objectLocation, locationID, err,
		)
	}

	freeIP, err := subnet.GetFreeIP(reservedIPs, true)
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to compute free ips for %s %s: %w", objectLocation, locationID, err,
		)
	}

	return freeIP, subnet.UUID, nil
}

var defaultHostNames = [52]string{
	"Einstein",
	"Gauss",
	"Newton",
	"Leibniz",
	"Euler",
	"Euclid",
	"Archimedes",
	"Turing",
	"Hilbert",
	"Pascal",
	"Fibonacci",
	"Pythagoras",
	"Descartes",
	"Fourier",
	"Fields",
	"Fermat",
	"Boole",
	"Kovalevskaya",
	"Lobachevsky",
	"Poincare",
	"Kolmogorov",
	"Bohr",
	"Lorentz",
	"Maxwell",
	"Planck",
	"Popov",
	"Rontgen",
	"Young",
	"Tesla",
	"Hawking",
	"Sakharov",
	"Dalton",
	"Faraday",
	"Curie",
	"Boyle",
	"Rutherford",
	"Joule",
	"Thomson",
	"Hertz",
	"Becquerel",
	"Landau",
	"Feynman",
	"Schrodinger",
	"Heisenberg",
	"Richmann",
	"Kapitsa",
	"Alferov",
	"Tamm",
	"Ioffe",
	"Vernadsky",
}

func resourceDedicatedServersServerV1GenerateHostNameIfNotPresented(schema *schema.ResourceData) string {
	osHostName, _ := schema.Get(dedicatedServersServerSchemaKeyOSHostName).(string)
	if osHostName == "" {
		osHostName = "tf-" + defaultHostNames[rand.IntN(len(defaultHostNames))]
	}

	return osHostName
}
