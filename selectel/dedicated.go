package selectel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used
	"net"
	"slices"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/servers"
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

func getDedicatedClient(d *schema.ResourceData, meta interface{}, withProjectScope bool) (*dedicated.ServiceClient, diag.Diagnostics) {
	config := meta.(*Config)

	var (
		selvpcClient *selvpcclient.Client
		err          error
	)

	if withProjectScope {
		projectID := d.Get(dedicatedServerSchemaKeyProjectID).(string)
		selvpcClient, err = config.GetSelVPCClientWithProjectScope(projectID)
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for dedicated servers api: %w", err))
		}
	} else {
		selvpcClient, err = config.GetSelVPCClient()
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("can't get selvpc client for dedicated servers api: %w", err))
		}
	}

	url := "https://api.selectel.ru/servers/v2"

	return dedicated.NewClientV2(selvpcClient.GetXAuthToken(), url), nil
}

// Partition config item types (API).
const (
	partitionTypeSoftRaid   = "soft_raid"
	partitionTypePartition  = "partition"
	partitionTypeLocalDrive = "local_drive"
	partitionTypeFilesystem = "filesystem"
)

type (
	PartitionsConfig struct {
		SoftRaidConfig []*SoftRaidConfigItem
		DiskPartitions []*DiskPartitionsItem
		DiskConfig     []*DiskConfigItem
	}

	SoftRaidConfigItem struct {
		Name     string
		Level    string
		DiskType string
		Count    int
	}

	DiskPartitionsItem struct {
		DiskName    string
		Mount       string
		Size        float64
		SizePercent float64
		Raid        string
		FSType      string
	}

	DiskConfigItem struct {
		Name     string
		DiskType string
	}
)

func (pc *PartitionsConfig) IsEmpty() bool {
	return len(pc.SoftRaidConfig) == 0 && len(pc.DiskPartitions) == 0 && len(pc.DiskConfig) == 0
}

const (
	mountBaseBoot = "/boot"
	mountBaseRoot = "/"
	mountBaseSwap = "swap"
)

func (pc *PartitionsConfig) CastToAPIPartitionsConfig(
	localDrives dedicated.LocalDrives,
	defaultPartitions []*dedicated.PartitionConfigItem,
) (dedicated.PartitionsConfig, error) {
	if err := pc.ensureDefaultConfig(localDrives, defaultPartitions); err != nil {
		return nil, err
	}

	res := buildBaseAPIConfig(localDrives)

	usedDrives := make(map[string]bool)

	raidMembers, err := pc.allocateSoftRaids(localDrives, usedDrives)
	if err != nil {
		return nil, err
	}

	diskNameToDrive, err := pc.allocateDiskConfigs(localDrives, usedDrives)
	if err != nil {
		return nil, err
	}

	if err = pc.ensureBootPartition(localDrives, defaultPartitions); err != nil {
		return nil, err
	}

	var (
		diskPartitions      = make([]*DiskPartitionsItem, 0, len(pc.DiskPartitions))
		nextPriorityByDrive = make(map[string]int)
	)

	for _, dp := range pc.DiskPartitions {
		isBase := slices.Contains([]string{
			mountBaseSwap, mountBaseRoot, mountBaseBoot,
		}, dp.Mount)

		if !isBase {
			err = pc.addDiskPartitionToAPIConfig(
				dp, localDrives, nextPriorityByDrive, res,
				raidMembers, diskNameToDrive,
			)
			if err != nil {
				return nil, err
			}

			continue
		}

		diskPartitions = append(diskPartitions, dp)
	}

	for _, dp := range diskPartitions {
		err = pc.addDiskPartitionToAPIConfig(
			dp, localDrives, nextPriorityByDrive, res,
			raidMembers, diskNameToDrive,
		)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (pc *PartitionsConfig) ensureDefaultConfig(
	localDrives dedicated.LocalDrives,
	defaultPartitions []*dedicated.PartitionConfigItem,
) error {
	if !pc.IsEmpty() {
		return nil
	}

	if len(localDrives) == 0 {
		return errors.New("local drives are required for automatic partitioning")
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

	return nil
}

func buildBaseAPIConfig(localDrives dedicated.LocalDrives) dedicated.PartitionsConfig {
	res := make(dedicated.PartitionsConfig)

	for ldID, ld := range localDrives {
		res[ldID] = &dedicated.PartitionConfigItem{
			Type: ld.Type,
			Match: &dedicated.PartitionConfigItemMatch{
				Size: ld.Match.Size,
				Type: ld.Match.Type,
			},
		}
	}

	return res
}

func (pc *PartitionsConfig) allocateSoftRaids(
	localDrives dedicated.LocalDrives,
	usedDrives map[string]bool,
) (map[string][]string, error) {
	raidMembers := make(map[string][]string)

	for _, sr := range pc.SoftRaidConfig {
		var candidates []string

		for ldID, ld := range localDrives {
			if usedDrives[ldID] {
				continue
			}
			if ld.Match.Type != sr.DiskType {
				continue
			}
			candidates = append(candidates, ldID)
		}

		if sr.Count > 0 && len(candidates) > sr.Count {
			candidates = candidates[:sr.Count]
		}

		if len(candidates) == 0 {
			return nil, fmt.Errorf("no drives for raid %s", sr.Name)
		}

		for _, id := range candidates {
			usedDrives[id] = true
		}

		raidMembers[sr.Name] = candidates
	}

	return raidMembers, nil
}

func (pc *PartitionsConfig) allocateDiskConfigs(
	localDrives dedicated.LocalDrives,
	usedDrives map[string]bool,
) (map[string]string, error) {
	result := make(map[string]string)

	for _, dc := range pc.DiskConfig {
		for ldID, ld := range localDrives {
			if usedDrives[ldID] {
				continue
			}
			if ld.Match.Type != dc.DiskType {
				continue
			}

			result[dc.Name] = ldID
			usedDrives[ldID] = true

			break
		}

		if result[dc.Name] == "" {
			return nil, fmt.Errorf("no free drive for disk_config %s", dc.Name)
		}
	}

	return result, nil
}

func (pc *PartitionsConfig) ensureBootPartition(
	localDrives dedicated.LocalDrives,
	defaultPartitions []*dedicated.PartitionConfigItem,
) error {
	if pc.ContainsBootPartition() {
		return nil
	}

	for _, dp := range defaultPartitions {
		if dp.Mount == mountBaseBoot {
			pc.DiskPartitions = append(pc.DiskPartitions, &DiskPartitionsItem{
				Raid:   pc.PickDefaultBootRaidName(localDrives),
				Mount:  dp.Mount,
				Size:   dp.Size,
				FSType: dp.FSType,
			})

			return nil
		}
	}

	return errors.New("can't find default partition for boot partition")
}

func (pc *PartitionsConfig) resolveDevicesForPartition(
	diskPartition *DiskPartitionsItem,
	localDrives dedicated.LocalDrives,
	raidMembers map[string][]string,
	diskNameToDrive map[string]string,
) (devices []string, raidLevel string, err error) {
	if diskPartition.Raid != "" {
		members := raidMembers[diskPartition.Raid]
		if len(members) == 0 {
			return nil, "", fmt.Errorf("raid %s has no devices", diskPartition.Raid)
		}
		devices = members

		for _, sr := range pc.SoftRaidConfig {
			if sr.Name == diskPartition.Raid {
				raidLevel = sr.Level

				break
			}
		}

		return devices, raidLevel, nil
	}

	if diskPartition.DiskName != "" {
		d := diskNameToDrive[diskPartition.DiskName]
		if d == "" {
			return nil, "", fmt.Errorf("disk %s not allocated", diskPartition.DiskName)
		}

		return []string{d}, "", nil
	}

	bestRatio := -1
	bestSize := -1
	bestID := ""
	for ldID, ld := range localDrives {
		if ld.Match == nil {
			continue
		}
		ratio := ld.SpeedRatio()
		size := ld.Match.Size
		if ratio > bestRatio || (ratio == bestRatio && size > bestSize) {
			bestRatio = ratio
			bestSize = size
			bestID = ldID
		}
	}
	if bestID != "" {
		devices = []string{bestID}
	}

	return devices, "", nil
}

func buildPartitionMembers(
	diskPartition *DiskPartitionsItem,
	devices []string,
	localDrives dedicated.LocalDrives,
	nextPriorityByDrive map[string]int,
	cfg dedicated.PartitionsConfig,
) ([]string, error) {
	members := make([]string, 0, len(devices))

	for _, ldID := range devices {
		ld := localDrives[ldID]

		id, err := uuid.GenerateUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate uuid for partition %s local drive %s: %w",
				diskPartition.Mount, ldID, err)
		}

		size := diskPartition.Size
		if diskPartition.SizePercent > 0 {
			baseSize := float64(ld.Match.Size)
			if baseSize <= 0 {
				return nil, fmt.Errorf("invalid local drive %s size: %d",
					ld.Match.Type, ld.Match.Size)
			}
			size = math.Round(baseSize * diskPartition.SizePercent / 100.0)
		}

		priority := nextPriorityByDrive[ldID]
		cfg[id] = &dedicated.PartitionConfigItem{
			Type:     "partition",
			Device:   ldID,
			Size:     size,
			Priority: &priority,
		}
		nextPriorityByDrive[ldID] = priority + 1
		members = append(members, id)
	}

	return members, nil
}

func (pc *PartitionsConfig) addDiskPartitionToAPIConfig(
	diskPartition *DiskPartitionsItem,
	localDrives dedicated.LocalDrives,
	nextPriorityByDrive map[string]int,
	cfg dedicated.PartitionsConfig,
	raidMembers map[string][]string,
	diskNameToDrive map[string]string,
) error {
	devices, raidLevel, err := pc.resolveDevicesForPartition(
		diskPartition, localDrives, raidMembers, diskNameToDrive,
	)
	if err != nil {
		return err
	}

	members, err := buildPartitionMembers(
		diskPartition, devices, localDrives, nextPriorityByDrive, cfg,
	)
	if err != nil {
		return err
	}

	var fsPartitionDeviceID string

	switch {
	case len(members) == 1:
		fsPartitionDeviceID = members[0]

	case len(members) > 1:
		srID, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate uuid for soft_raid %s: %w",
				diskPartition.Mount, err)
		}
		cfg[srID] = &dedicated.PartitionConfigItem{
			Type:    "soft_raid",
			Members: members,
			Level:   raidLevel,
		}
		fsPartitionDeviceID = srID

	default:
		return errors.New("no devices for " + diskPartition.Mount)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return fmt.Errorf("failed to generate uuid for filesystem partition %s: %w",
			diskPartition.Mount, err)
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

	cfg[id] = &dedicated.PartitionConfigItem{
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

func (pc *PartitionsConfig) PickDefaultBootRaidName(localDrives dedicated.LocalDrives) string {
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

func resourceDedicatedServerV1ReadPartitionsConfig(d *schema.ResourceData) (*PartitionsConfig, error) {
	res := new(PartitionsConfig)

	v, ok := d.GetOk(dedicatedServerSchemaKeyOSPartitionsConfig)
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
	res.SoftRaidConfig, err = resourceDedicatedServerV1ReadPartitionsConfigSoftRaid(partitionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read soft raid configuration: %w", err)
	}

	res.DiskPartitions, err = resourceDedicatedServerV1ReadPartitionsConfigDiskPartitions(partitionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read disk partitions configuration: %w", err)
	}

	res.DiskConfig, err = resourceDedicatedServerV1ReadPartitionsConfigDiskConfig(partitionsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read disk configuration: %w", err)
	}

	return res, nil
}

type RAIDLevel string

const (
	dedicatedServerRaid0Level  RAIDLevel = "raid0"
	dedicatedServerRaid1Level  RAIDLevel = "raid1"
	dedicatedServerRaid10Level RAIDLevel = "raid10"
)

func (rl RAIDLevel) MinDiskCount() (int, error) {
	switch rl {
	case dedicatedServerRaid0Level, dedicatedServerRaid1Level:
		return 2, nil
	case dedicatedServerRaid10Level:
		return 4, nil
	default:
		return 0, fmt.Errorf("unsupported raid level %s", rl)
	}
}

func resourceDedicatedServerV1ReadPartitionsConfigSoftRaid(partitionsConfig map[string]interface{}) ([]*SoftRaidConfigItem, error) {
	srCfgRaw, ok := partitionsConfig[dedicatedServerSchemaKeySoftRaidConfig].([]interface{})
	if !ok {
		srCfgRaw = make([]interface{}, 0)
	}

	res := make([]*SoftRaidConfigItem, 0, len(srCfgRaw))

	for idx, itemRaw := range srCfgRaw {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d] has unexpected type", idx)
		}

		name, ok := item[dedicatedServerSchemaKeyName].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].name has unexpected type", idx)
		}

		levelRaw, ok := item[dedicatedServerSchemaKeyLevel].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].level has unexpected type", idx)
		}

		diskType, ok := item[dedicatedServerSchemaKeyDiskType].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].disk_type has unexpected type", idx)
		}

		count, ok := item[dedicatedServerSchemaKeyDiskCount].(int)
		if !ok {
			return nil, fmt.Errorf("partitions_config.soft_raid_config[%d].count has unexpected type", idx)
		}

		level := RAIDLevel(levelRaw)

		minCount, err := level.MinDiskCount()
		if err != nil {
			return nil, fmt.Errorf(
				"partitions_config.soft_raid_config[%d]: %w",
				idx, err,
			)
		}

		if count == 0 {
			count = minCount
		} else if count < minCount {
			return nil, fmt.Errorf(
				"partitions_config.soft_raid_config[%d].count must be >= %d for %s, got %d",
				idx, minCount, level, count,
			)
		}

		res = append(res, &SoftRaidConfigItem{
			Name:     name,
			Level:    string(level),
			DiskType: diskType,
			Count:    count,
		})
	}

	return res, nil
}

func resourceDedicatedServerV1ReadPartitionsConfigDiskPartitions(partitionsConfig map[string]interface{}) ([]*DiskPartitionsItem, error) {
	dpCfgRaw, ok := partitionsConfig[dedicatedServerSchemaKeyDiskPartitions].([]interface{})
	if !ok {
		dpCfgRaw = make([]interface{}, 0)
	}

	res := make([]*DiskPartitionsItem, 0, len(dpCfgRaw))

	for idx, itemRaw := range dpCfgRaw {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d] has unexpected type", idx)
		}

		diskName, ok := item[dedicatedServerSchemaKeyDiskName].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d].disk_name has unexpected type", idx)
		}

		mount, ok := item[dedicatedServerSchemaKeyMount].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d].mount has unexpected type", idx)
		}

		size := 0.0

		sizeRaw, hasSizeRaw := item[dedicatedServerSchemaKeySize]
		if hasSizeRaw {
			size, ok = sizeRaw.(float64)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].size has unexpected type", idx)
			}
		}

		sizePercent := 0.0

		sizePercentRaw, hasSizePercentRaw := item[dedicatedServerSchemaKeySizePercent]
		if hasSizePercentRaw {
			sizePercent, ok = sizePercentRaw.(float64)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].size_percent has unexpected type", idx)
			}
		}

		if (size != 0 && sizePercent != 0) || (size == 0 && sizePercent == 0) {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d]: size or size_percent must be presented once", idx)
		}

		raid, ok := item[dedicatedServerSchemaKeyRaid].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_partitions[%d].raid has unexpected type", idx)
		}

		fsTypeRaw, ok := item[dedicatedServerSchemaKeyFSType]

		fsType := ""
		if ok {
			fsType, ok = fsTypeRaw.(string)
			if !ok {
				return nil, fmt.Errorf("partitions_config.disk_partitions[%d].fsType has unexpected type", idx)
			}
		}

		res = append(res, &DiskPartitionsItem{
			DiskName:    diskName,
			Mount:       mount,
			Size:        size,
			SizePercent: sizePercent,
			Raid:        raid,
			FSType:      fsType,
		})
	}

	return res, nil
}

func resourceDedicatedServerV1ReadPartitionsConfigDiskConfig(partitionsConfig map[string]any) ([]*DiskConfigItem, error) {
	dcCfgRaw, ok := partitionsConfig[dedicatedServerSchemaKeyDiskConfig].([]interface{})
	if !ok {
		dcCfgRaw = make([]any, 0)
	}

	res := make([]*DiskConfigItem, 0, len(dcCfgRaw))

	for idx, itemRaw := range dcCfgRaw {
		item, ok := itemRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_config[%d] has unexpected type", idx)
		}

		name, ok := item[dedicatedServerSchemaKeyName].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_config[%d].name has unexpected type", idx)
		}

		diskType, ok := item[dedicatedServerSchemaKeyDiskType].(string)
		if !ok {
			return nil, fmt.Errorf("partitions_config.disk_config[%d].disk_type has unexpected type", idx)
		}

		res = append(res, &DiskConfigItem{
			Name:     name,
			DiskType: diskType,
		})
	}

	return res, nil
}

func resourceDedicatedServerV1GetFreePublicIPs(
	ctx context.Context, cl *dedicated.ServiceClient, locationID, subnetID string,
) (net.IP, error) {
	subnet, _, err := cl.NetworkSubnet(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get subnets for %s %s: %w", objectLocation, locationID, err,
		)
	}

	nets, _, err := cl.Networks(ctx, locationID, dedicated.NetworkTypeInet, "")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get %s networks for %s %s: %w", dedicated.NetworkTypeInet, objectLocation, locationID, err,
		)
	}

	nets = nets.FilterByTelematicsTypeHosting()

	netsContainSubnet := slices.ContainsFunc(nets, func(n *dedicated.Network) bool {
		return subnet.NetworkUUID == n.UUID
	})

	if !netsContainSubnet || subnet.Free == 0 {
		return nil, fmt.Errorf(
			"subnet %s is not suitable for allocating ip", subnetID,
		)
	}

	reservedIPs, _, err := cl.NetworkReservedIPs(ctx, locationID, "")
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

func resourceDedicatedServerV1GetFreePrivateIPs(
	ctx context.Context, cl *dedicated.ServiceClient, locationID, subnetStr string,
) (ip net.IP, subnetID string, err error) {
	nets, _, err := cl.Networks(ctx, locationID, dedicated.NetworkTypeLocal, "")
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to get %s networks for %s %s: %w", dedicated.NetworkTypeInet, objectLocation, locationID, err,
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

func resourceDedicatedServerV1GenerateHostNameIfNotPresented(schema *schema.ResourceData) string {
	osHostName, _ := schema.Get(dedicatedServerSchemaKeyOSHostName).(string)
	if osHostName == "" {
		osHostName = "tf-" + defaultHostNames[rand.IntN(len(defaultHostNames))]
	}

	return osHostName
}

func apiPartitionsConfigToSchema(
	apiCfg dedicated.PartitionsConfig,
	existingNamesByType map[string][]string,
	existingMounts []string,
	existingDiskNameByMount map[string]string,
	existingDiskConfigOrder []string,
	existingRaidNamesByKey map[string][]string,
	existingRaidConfigOrder []string,
) ([]map[string]any, error) {
	if len(apiCfg) == 0 {
		return nil, nil
	}

	var (
		raidNamesByID = make(map[string]string)
		diskNamesByID = make(map[string]string)
		diskInRaid    = make(map[string]bool)
	)

	softRaids := buildSoftRaids(apiCfg, raidNamesByID, diskInRaid, existingRaidNamesByKey, existingRaidConfigOrder)
	diskConfigs := buildDiskConfigs(apiCfg, diskNamesByID, diskInRaid, existingNamesByType, existingDiskNameByMount, existingDiskConfigOrder)
	diskPartitions := buildDiskPartitions(apiCfg, raidNamesByID, diskNamesByID, existingMounts)

	result := map[string]any{
		dedicatedServerSchemaKeySoftRaidConfig: softRaids,
		dedicatedServerSchemaKeyDiskPartitions: diskPartitions,
	}

	if len(diskConfigs) > 0 {
		result[dedicatedServerSchemaKeyDiskConfig] = diskConfigs
	}

	return []map[string]any{result}, nil
}

// buildSoftRaids creates soft_raid_config entries from API config.
// existingRaidNamesByKey maps "level|diskType" → ordered slice of prior names; when
// non-empty the first unused name for each key is reused so user-defined labels are
// preserved across Read cycles instead of regenerating "new-raid1" etc.
// existingRaidConfigOrder lists RAID names in the order they appear in prior state;
// used to sort the result to match the user's TypeList order and avoid positional diffs.
func buildSoftRaids( //nolint:gocognit
	apiCfg dedicated.PartitionsConfig, raidNamesByID map[string]string, diskInRaid map[string]bool,
	existingRaidNamesByKey map[string][]string,
	existingRaidConfigOrder []string,
) []map[string]any {
	softRaids := make([]map[string]any, 0)

	// First pass: group RAID arrays by disk type and level
	raidKeyInfo := make(map[string]struct {
		level       string
		diskType    string
		uniqueDisks int
		diskSet     map[string]bool
		members     int
	})

	for _, item := range apiCfg {
		if item.Type != partitionTypeSoftRaid {
			continue
		}

		diskType := getDiskTypeForRaid(apiCfg, item.Members)
		key := item.Level + "|" + diskType

		// Count unique physical disks for this RAID
		diskSet := make(map[string]bool)
		for _, memberID := range item.Members {
			member, exists := apiCfg[memberID]
			if exists && member.Type == partitionTypePartition && member.Device != "" {
				device, exists := apiCfg[member.Device]
				if exists && device.Type == partitionTypeLocalDrive {
					diskSet[member.Device] = true
				}
			}
		}

		info := raidKeyInfo[key]
		info.level = item.Level
		info.diskType = diskType
		info.members += len(item.Members)
		if info.diskSet == nil {
			info.diskSet = make(map[string]bool)
		}
		for diskID := range diskSet {
			info.diskSet[diskID] = true
		}
		info.uniqueDisks = len(info.diskSet)
		raidKeyInfo[key] = info
	}

	// Second pass: create RAID configs, preserving user-defined names from prior state.
	raidKeys := make([]string, 0, len(raidKeyInfo))
	for k := range raidKeyInfo {
		raidKeys = append(raidKeys, k)
	}
	slices.Sort(raidKeys)

	raidNameIdx := make(map[string]int)
	for _, raidKey := range raidKeys {
		info := raidKeyInfo[raidKey]

		var raidName string
		if names := existingRaidNamesByKey[raidKey]; len(names) > 0 {
			idx := raidNameIdx[raidKey]
			if idx < len(names) {
				raidName = names[idx]
			}
			raidNameIdx[raidKey] = idx + 1
		}

		if raidName == "" {
			// No prior state — generate a stable fallback name.
			genKey := "gen|" + info.level
			if len(raidKeyInfo) > 1 {
				idx := raidNameIdx[genKey]
				raidName = fmt.Sprintf("new-%s-%d", info.level, idx)
				raidNameIdx[genKey] = idx + 1
			} else {
				raidName = "new-" + info.level
			}
		}

		// Store RAID name for all RAID items with this key
		for id, item := range apiCfg {
			if item.Type != partitionTypeSoftRaid {
				continue
			}

			itemDiskType := getDiskTypeForRaid(apiCfg, item.Members)
			if itemDiskType != info.diskType || item.Level != info.level {
				continue
			}

			raidNamesByID[id] = raidName
			// Mark member disks as being in RAID
			for _, memberID := range item.Members {
				member, exists := apiCfg[memberID]
				if exists && member.Type == partitionTypePartition && member.Device != "" {
					diskInRaid[member.Device] = true
				}
			}
		}

		softRaid := map[string]any{
			dedicatedServerSchemaKeyName:      raidName,
			dedicatedServerSchemaKeyLevel:     info.level,
			dedicatedServerSchemaKeyDiskType:  info.diskType,
			dedicatedServerSchemaKeyDiskCount: info.uniqueDisks,
		}
		softRaids = append(softRaids, softRaid)
	}

	// Sort soft_raids: use prior-state order when available, otherwise alphabetically.
	if len(existingRaidConfigOrder) > 0 {
		orderIdx := make(map[string]int, len(existingRaidConfigOrder))
		for i, name := range existingRaidConfigOrder {
			orderIdx[name] = i
		}
		slices.SortFunc(softRaids, func(a, b map[string]any) int {
			nameA, _ := a[dedicatedServerSchemaKeyName].(string)
			nameB, _ := b[dedicatedServerSchemaKeyName].(string)
			posA, okA := orderIdx[nameA]
			posB, okB := orderIdx[nameB]
			if !okA {
				posA = len(existingRaidConfigOrder)
			}
			if !okB {
				posB = len(existingRaidConfigOrder)
			}
			if posA != posB {
				return posA - posB
			}

			return strings.Compare(nameA, nameB)
		})
	} else {
		slices.SortFunc(softRaids, func(a, b map[string]any) int {
			nameA, _ := a[dedicatedServerSchemaKeyName].(string)
			nameB, _ := b[dedicatedServerSchemaKeyName].(string)
			return strings.Compare(nameA, nameB)
		})
	}

	return softRaids
}

// findMountsForDrive returns the filesystem mount points served by a given local drive.
// It traces: driveID → partition items where Device=driveID → filesystem items where Device=partitionID.
func findMountsForDrive(apiCfg dedicated.PartitionsConfig, driveID string) []string {
	var mounts []string
	for partID, partItem := range apiCfg {
		if partItem.Type != partitionTypePartition || partItem.Device != driveID {
			continue
		}
		for _, fsItem := range apiCfg {
			if fsItem.Type == partitionTypeFilesystem && fsItem.Device == partID {
				mounts = append(mounts, fsItem.Mount)
			}
		}
	}

	return mounts
}

// resolveDiskName returns the disk name for a given drive ID, or (_, false) if the drive
// should be skipped. It updates typeNameIndex and seenNames as a side-effect.
func resolveDiskName(
	apiCfg dedicated.PartitionsConfig,
	id string,
	diskType string,
	existingDiskNameByMount map[string]string,
	existingNamesByType map[string][]string,
	typeNameIndex map[string]int,
	seenNames map[string]bool,
) (name string, ok bool) {
	if len(existingDiskNameByMount) > 0 {
		for _, mount := range findMountsForDrive(apiCfg, id) {
			if n, found := existingDiskNameByMount[mount]; found && n != "" {
				name = n
				break
			}
		}
		if name == "" || seenNames[name] {
			return "", false
		}
		seenNames[name] = true

		return name, true
	}

	mounts := findMountsForDrive(apiCfg, id)
	if len(mounts) == 0 {
		return "", false
	}

	idx := typeNameIndex[diskType]
	names := existingNamesByType[diskType]
	if idx < len(names) {
		name = names[idx]
	} else {
		name = generateDiskName(diskType, idx)
	}
	typeNameIndex[diskType]++

	if seenNames[name] {
		return "", false
	}
	seenNames[name] = true

	return name, true
}

// buildDiskConfigs creates disk_config entries for drives not in RAID.
// One entry is created per physical drive that has at least one filesystem mount.
// When existingDiskNameByMount is provided the function traces each drive through its
// partitions to a filesystem mount and looks up the user-assigned name; drives with no
// matching mounts are skipped (// drive not in prior config; skip to avoid phantom entries).
// When existingDiskNameByMount is empty the fallback assigns names from existingNamesByType
// by position within each type, or generates a name. Drives with no mounts are always skipped.
// When existingDiskConfigOrder is provided disk_configs are sorted to match that order;
// otherwise sorted alphabetically by name.
func buildDiskConfigs(
	apiCfg dedicated.PartitionsConfig,
	diskNamesByID map[string]string,
	diskInRaid map[string]bool,
	existingNamesByType map[string][]string,
	existingDiskNameByMount map[string]string,
	existingDiskConfigOrder []string,
) []map[string]any {
	diskConfigs := make([]map[string]any, 0)
	typeNameIndex := make(map[string]int)
	seenNames := make(map[string]bool)

	for id, item := range apiCfg {
		if item.Type != partitionTypeLocalDrive || diskInRaid[id] {
			continue
		}

		diskType := item.Match.Type
		diskName, ok := resolveDiskName(
			apiCfg, id, diskType,
			existingDiskNameByMount, existingNamesByType,
			typeNameIndex, seenNames,
		)
		if !ok {
			continue
		}

		diskNamesByID[id] = diskName
		diskConfigs = append(diskConfigs, map[string]any{
			dedicatedServerSchemaKeyName:     diskName,
			dedicatedServerSchemaKeyDiskType: diskType,
		})
	}

	if len(existingDiskConfigOrder) > 0 {
		orderIdx := make(map[string]int, len(existingDiskConfigOrder))
		for i, name := range existingDiskConfigOrder {
			orderIdx[name] = i
		}
		slices.SortFunc(diskConfigs, func(a, b map[string]any) int {
			nameA, _ := a[dedicatedServerSchemaKeyName].(string)
			nameB, _ := b[dedicatedServerSchemaKeyName].(string)
			posA, okA := orderIdx[nameA]
			posB, okB := orderIdx[nameB]
			if !okA {
				posA = len(existingDiskConfigOrder)
			}
			if !okB {
				posB = len(existingDiskConfigOrder)
			}
			if posA != posB {
				return posA - posB
			}

			return strings.Compare(nameA, nameB)
		})
	} else {
		slices.SortFunc(diskConfigs, func(a, b map[string]any) int {
			nameA, _ := a[dedicatedServerSchemaKeyName].(string)
			nameB, _ := b[dedicatedServerSchemaKeyName].(string)

			return strings.Compare(nameA, nameB)
		})
	}

	return diskConfigs
}

func mountSortPriority(mount string) int {
	switch mount {
	case mountBaseBoot:
		return 0
	case mountBaseSwap:
		return 1
	case mountBaseRoot:
		return 2
	default:
		return 3
	}
}

// buildDiskPartitions creates disk_partitions entries from filesystem items.
// When existingMounts is non-empty only partitions whose mount appears in that set are returned.
func buildDiskPartitions(
	apiCfg dedicated.PartitionsConfig, raidNamesByID, diskNamesByID map[string]string,
	existingMounts []string,
) []map[string]any {
	diskPartitions := make([]map[string]any, 0)

	var mountFilter map[string]bool
	if len(existingMounts) > 0 {
		mountFilter = make(map[string]bool, len(existingMounts))
		for _, m := range existingMounts {
			mountFilter[m] = true
		}
	}

	for _, item := range apiCfg {
		if item.Type != partitionTypeFilesystem {
			continue
		}

		if mountFilter != nil && !mountFilter[item.Mount] {
			continue
		}

		partition := map[string]any{
			dedicatedServerSchemaKeyMount: item.Mount,
		}

		if item.FSType != "" {
			partition[dedicatedServerSchemaKeyFSType] = item.FSType
		}

		if item.Device != "" {
			applyDeviceConfig(apiCfg, item.Device, partition, raidNamesByID, diskNamesByID)
		}

		diskPartitions = append(diskPartitions, partition)
	}

	if len(existingMounts) > 0 {
		mountIdx := make(map[string]int, len(existingMounts))
		for i, m := range existingMounts {
			mountIdx[m] = i
		}
		slices.SortFunc(diskPartitions, func(a, b map[string]any) int {
			mountA, _ := a[dedicatedServerSchemaKeyMount].(string)
			mountB, _ := b[dedicatedServerSchemaKeyMount].(string)
			posA, okA := mountIdx[mountA]
			posB, okB := mountIdx[mountB]
			if !okA {
				posA = len(existingMounts)
			}
			if !okB {
				posB = len(existingMounts)
			}
			if posA != posB {
				return posA - posB
			}

			return strings.Compare(mountA, mountB)
		})
	} else {
		slices.SortFunc(diskPartitions, func(a, b map[string]any) int {
			mountA, _ := a[dedicatedServerSchemaKeyMount].(string)
			mountB, _ := b[dedicatedServerSchemaKeyMount].(string)
			prioA := mountSortPriority(mountA)
			prioB := mountSortPriority(mountB)
			if prioA != prioB {
				return prioA - prioB
			}

			return strings.Compare(mountA, mountB)
		})
	}

	return diskPartitions
}

// applyDeviceConfig applies disk_name or raid configuration based on device type.
func applyDeviceConfig(
	apiCfg dedicated.PartitionsConfig, deviceID string, partition map[string]any,
	raidNamesByID, diskNamesByID map[string]string,
) {
	deviceItem, exists := apiCfg[deviceID]
	if !exists {
		return
	}

	switch deviceItem.Type {
	case partitionTypeSoftRaid:
		if raidName, ok := raidNamesByID[deviceID]; ok {
			partition[dedicatedServerSchemaKeyRaid] = raidName
			size := findSizeForRaidMember(apiCfg, deviceID)
			if size != 0 {
				partition[dedicatedServerSchemaKeySize] = size
			}
		}

	case partitionTypeLocalDrive:
		if diskName, ok := diskNamesByID[deviceID]; ok {
			partition[dedicatedServerSchemaKeyDiskName] = diskName
		}

	case partitionTypePartition:
		rootDiskName := findRootDiskName(apiCfg, deviceID, diskNamesByID)
		raidName := findRaidForPartition(apiCfg, deviceID, raidNamesByID)

		if raidName != "" {
			partition[dedicatedServerSchemaKeyRaid] = raidName
		} else if rootDiskName != "" {
			partition[dedicatedServerSchemaKeyDiskName] = rootDiskName
		}

		if deviceItem.Size != 0 {
			partition[dedicatedServerSchemaKeySize] = deviceItem.Size
		}
	}
}

// findRootDiskName traverses partition chain to find the root disk name.
func findRootDiskName(apiCfg dedicated.PartitionsConfig, deviceID string, diskNamesByID map[string]string) string {
	visited := make(map[string]bool)
	currentID := deviceID

	for !visited[currentID] {
		visited[currentID] = true

		item, exists := apiCfg[currentID]
		if !exists {
			break
		}

		switch item.Type {
		case partitionTypeLocalDrive:
			if diskName, ok := diskNamesByID[currentID]; ok {
				return diskName
			}

			return ""
		case partitionTypePartition:
			currentID = item.Device
		default:
			break
		}
	}

	return ""
}

// findRaidForPartition finds if a partition belongs to a RAID array.
func findRaidForPartition(apiCfg dedicated.PartitionsConfig, partitionID string, raidNamesByID map[string]string) string {
	// First check if the partition itself is a member of a RAID
	for raidID, item := range apiCfg {
		if item.Type != partitionTypeSoftRaid {
			continue
		}

		for _, memberID := range item.Members {
			if memberID == partitionID {
				return raidNamesByID[raidID]
			}
		}
	}

	// Then check if the partition's device is a RAID
	item, exists := apiCfg[partitionID]
	if exists && item.Type == partitionTypeSoftRaid {
		return raidNamesByID[partitionID]
	}

	// Traverse the chain to find RAID
	visited := make(map[string]bool)
	currentID := partitionID
	for !visited[currentID] {
		visited[currentID] = true

		currItem, exists := apiCfg[currentID]
		if !exists {
			break
		}

		switch currItem.Type {
		case partitionTypeSoftRaid:
			return raidNamesByID[currentID]
		case partitionTypePartition:
			currentID = currItem.Device
		default:
			break
		}
	}

	return ""
}

// findSizeForRaidMember finds the size for a specific filesystem within a RAID.
// For RAID0/RAID10 sizes are multiplied by member count, for RAID1 size is as-is.
func findSizeForRaidMember(apiCfg dedicated.PartitionsConfig, raidID string) float64 {
	raidItem, exists := apiCfg[raidID]
	if !exists || raidItem.Type != "soft_raid" {
		return 0
	}

	// Get size from any partition member
	var memberSize float64
	for _, memberID := range raidItem.Members {
		member, exists := apiCfg[memberID]
		if exists && member.Type == "partition" && member.Size != 0 {
			memberSize = member.Size
			break
		}
	}

	// Don't multiply negative sizes (they mean "all remaining space")
	if memberSize < 0 {
		return memberSize
	}

	// Apply RAID multiplier based on level
	// RAID0: sizes are striped across all disks (total = member_size * count)
	// RAID1: mirrored (total = member_size)
	// RAID10: striped mirrors (total = member_size * count / 2)
	memberCount := len(raidItem.Members)
	switch RAIDLevel(raidItem.Level) {
	case dedicatedServerRaid0Level:
		return memberSize * float64(memberCount)
	case dedicatedServerRaid1Level, dedicatedServerRaid10Level:
		return memberSize
	default:
		return memberSize
	}
}

// generateDiskName generates a name for disk based on type and index.
func generateDiskName(diskType string, index int) string {
	diskName := strings.ToLower(strings.ReplaceAll(diskType, " ", "-"))
	if index > 0 {
		return fmt.Sprintf("disk-%s-%d", diskName, index)
	}

	return fmt.Sprintf("disk-%s", diskName)
}

// getDiskTypeForRaid determines the disk type for a RAID based on its members.
func getDiskTypeForRaid(apiCfg dedicated.PartitionsConfig, memberIDs []string) string {
	for _, memberID := range memberIDs {
		member, exists := apiCfg[memberID]
		if !exists || member.Type != "partition" || member.Device == "" {
			continue
		}

		device, exists := apiCfg[member.Device]
		if !exists || device.Type != "local_drive" {
			continue
		}

		return device.Match.Type
	}

	return ""
}
