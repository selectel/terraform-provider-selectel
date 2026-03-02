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

func (pc *PartitionsConfig) addDiskPartitionToAPIConfig(
	diskPartition *DiskPartitionsItem,
	localDrives dedicated.LocalDrives,
	nextPriorityByDrive map[string]int,
	cfg dedicated.PartitionsConfig,
	raidMembers map[string][]string,
	diskNameToDrive map[string]string,
) error {
	var devices []string
	var raidLevel string

	// RAID source
	if diskPartition.Raid != "" {
		members := raidMembers[diskPartition.Raid]
		if len(members) == 0 {
			return fmt.Errorf("raid %s has no devices", diskPartition.Raid)
		}
		devices = members

		for _, sr := range pc.SoftRaidConfig {
			if sr.Name == diskPartition.Raid {
				raidLevel = sr.Level

				break
			}
		}
	}

	// disk_config source
	if diskPartition.DiskName != "" {
		d := diskNameToDrive[diskPartition.DiskName]
		if d == "" {
			return fmt.Errorf("disk %s not allocated", diskPartition.DiskName)
		}
		devices = []string{d}
	}

	// fallback
	if len(devices) == 0 {
		for ldID := range localDrives {
			devices = []string{ldID}

			break
		}
	}

	members := make([]string, 0, len(devices))

	for _, ldID := range devices {
		ld := localDrives[ldID]

		id, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate uuid for partition %s local drive %s: %w",
				diskPartition.Mount, ldID, err)
		}

		size := diskPartition.Size
		if diskPartition.SizePercent > 0 {
			baseSize := float64(ld.Match.Size)
			if baseSize <= 0 {
				return fmt.Errorf("invalid local drive %s size: %d",
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
