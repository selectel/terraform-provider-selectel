package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dedicated "github.com/selectel/dedicated-go/v2/pkg/v2"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/servers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServersMapsFromStructs(t *testing.T) {
	serverStructs := []servers.Server{
		{
			ID:     "a208023f-69fe-4a9e-8285-dd44e94a854a",
			Name:   "fake",
			Status: "ACTIVE",
		},
	}
	expectedServersMaps := []map[string]interface{}{
		{
			"id":     "a208023f-69fe-4a9e-8285-dd44e94a854a",
			"name":   "fake",
			"status": "ACTIVE",
		},
	}

	actualServersMaps := serversMapsFromStructs(serverStructs)

	assert.Equal(t, expectedServersMaps, actualServersMaps)
}

func newTestDedicatedAPIClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) *dedicated.ServiceClient {
	config := testAccProvider.Meta().(*Config)

	var projectID string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		panic("can't get selvpc client for dedicated servers acc tests: " + err.Error())
	}

	url := "https://api.selectel.ru/servers/v2"

	return dedicated.NewClientV2(selvpcClient.GetXAuthToken(), url)
}

func TestPartitionsConfig_IsEmpty(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		pc := &PartitionsConfig{}
		assert.True(t, pc.IsEmpty())
	})
	t.Run("FalseWithSoftRaid", func(t *testing.T) {
		pc := &PartitionsConfig{SoftRaidConfig: []*SoftRaidConfigItem{{}}}
		assert.False(t, pc.IsEmpty())
	})
	t.Run("FalseWithDiskPartitions", func(t *testing.T) {
		pc := &PartitionsConfig{DiskPartitions: []*DiskPartitionsItem{{}}}
		assert.False(t, pc.IsEmpty())
	})
}

func TestPartitionsConfig_ContainsBootPartition(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		pc := &PartitionsConfig{DiskPartitions: []*DiskPartitionsItem{{Mount: "/boot"}}}
		assert.True(t, pc.ContainsBootPartition())
	})
	t.Run("False", func(t *testing.T) {
		pc := &PartitionsConfig{DiskPartitions: []*DiskPartitionsItem{{Mount: "/"}}}
		assert.False(t, pc.ContainsBootPartition())
	})
	t.Run("FalseWhenEmpty", func(t *testing.T) {
		pc := &PartitionsConfig{}
		assert.False(t, pc.ContainsBootPartition())
	})
}

func TestPartitionsConfig_CastToAPIPartitionsConfig(t *testing.T) {
	localDrives := dedicated.LocalDrives{
		"drive-ssd-1": {
			Type: "drive",
			Match: &dedicated.LocalDriveMatch{
				Size: 1000,
				Type: "SSD",
			},
		},
		"drive-ssd-2": {
			Type: "drive",
			Match: &dedicated.LocalDriveMatch{
				Size: 1000,
				Type: "SSD",
			},
		},
		"drive-hdd-1": {
			Type: "drive",
			Match: &dedicated.LocalDriveMatch{
				Size: 2000,
				Type: "SATA",
			},
		},
	}

	defaultPartitions := []*dedicated.PartitionConfigItem{
		{Mount: "/boot", Size: 1, FSType: "ext3"},
		{Mount: "/", Size: -1, FSType: "ext4"},
		{Mount: "swap", Size: 12, FSType: "swap"},
	}

	findItemsByType := func(cfg dedicated.PartitionsConfig, itemType string) []*dedicated.PartitionConfigItem {
		var items []*dedicated.PartitionConfigItem
		for _, item := range cfg {
			if item.Type == itemType {
				items = append(items, item)
			}
		}

		return items
	}

	findFSByMount := func(cfg dedicated.PartitionsConfig, mount string) *dedicated.PartitionConfigItem {
		for _, item := range cfg {
			if item.Type == "filesystem" && item.Mount == mount {
				return item
			}
		}

		return nil
	}

	t.Run("AutomaticSuccess", func(t *testing.T) {
		pc := &PartitionsConfig{}
		apiConfig, err := pc.CastToAPIPartitionsConfig(localDrives, defaultPartitions)
		require.NoError(t, err)

		drives := findItemsByType(apiConfig, "drive")
		assert.Len(t, drives, 3)

		softRaids := findItemsByType(apiConfig, "soft_raid")
		require.Len(t, softRaids, 3)
		for _, sr := range softRaids {
			assert.Equal(t, "raid1", sr.Level)
			assert.Len(t, sr.Members, 2)
		}

		p := findItemsByType(apiConfig, "partition")
		require.Len(t, p, 6)

		bootFS := findFSByMount(apiConfig, "/boot")
		require.NotNil(t, bootFS)
		assert.Equal(t, "ext3", bootFS.FSType)

		rootFS := findFSByMount(apiConfig, "/")
		require.NotNil(t, rootFS)
		assert.Equal(t, "ext4", rootFS.FSType)

		swapFS := findFSByMount(apiConfig, "swap")
		require.NotNil(t, swapFS)
		assert.Equal(t, "swap", swapFS.FSType)
	})

	t.Run("AutomaticFailNoDrives", func(t *testing.T) {
		pc := &PartitionsConfig{}
		_, err := pc.CastToAPIPartitionsConfig(nil, defaultPartitions)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "local drives are required for automatic partitioning")
	})

	t.Run("ManualSuccessSimple", func(t *testing.T) {
		pc := &PartitionsConfig{
			SoftRaidConfig: []*SoftRaidConfigItem{
				{Name: "ssd-raid", Level: "raid1", DiskType: "SSD"},
			},
			DiskPartitions: []*DiskPartitionsItem{
				{Mount: "/", Size: -1, Raid: "ssd-raid", FSType: "xfs"},
				{Mount: "backup", Size: -1, Raid: "ssd-raid", FSType: "xfs"},
			},
		}
		apiConfig, err := pc.CastToAPIPartitionsConfig(localDrives, defaultPartitions)
		require.NoError(t, err)

		bootFS := findFSByMount(apiConfig, "/boot")
		require.NotNil(t, bootFS)
		assert.Equal(t, "ext3", bootFS.FSType)

		rootFS := findFSByMount(apiConfig, "/")
		require.NotNil(t, rootFS)
		assert.Equal(t, "xfs", rootFS.FSType)

		backupFS := findFSByMount(apiConfig, "backup")
		require.NotNil(t, backupFS)
		assert.Equal(t, "xfs", backupFS.FSType)

		softRaids := findItemsByType(apiConfig, "soft_raid")
		require.Len(t, softRaids, 3)
		assert.Equal(t, "raid1", softRaids[0].Level)
	})

	t.Run("ManualFailRaidNotFound", func(t *testing.T) {
		pc := &PartitionsConfig{
			DiskPartitions: []*DiskPartitionsItem{
				{Mount: "/", Size: -1, Raid: "non-existent-raid"},
			},
		}
		_, err := pc.CastToAPIPartitionsConfig(localDrives, defaultPartitions)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "raid non-existent-raid has no devices")
	})

	t.Run("ManualFailNoDefaultBoot", func(t *testing.T) {
		pc := &PartitionsConfig{
			SoftRaidConfig: []*SoftRaidConfigItem{{Name: "raid", Level: "raid1", DiskType: "SSD"}},
			DiskPartitions: []*DiskPartitionsItem{{Mount: "/", Size: -1, Raid: "raid"}},
		}
		_, err := pc.CastToAPIPartitionsConfig(localDrives, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't find default partition for boot partition")
	})

	t.Run("ManualFailNoDisksForRaid", func(t *testing.T) {
		pc := &PartitionsConfig{
			SoftRaidConfig: []*SoftRaidConfigItem{{Name: "raid", Level: "raid1", DiskType: "NVME"}},
			DiskPartitions: []*DiskPartitionsItem{{Mount: "/", Size: -1, Raid: "raid"}},
		}
		_, err := pc.CastToAPIPartitionsConfig(localDrives, defaultPartitions)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no drives for raid raid")
	})
}

func TestApiPartitionsConfigToSchema(t *testing.T) {
	t.Run("SimpleRaidConfig", func(t *testing.T) {
		apiCfg := dedicated.PartitionsConfig{
			"d894a3f8-a17f-c141-76c4-e4ffdd240c56": {
				Type: "soft_raid",
				Members: []string{
					"695ea5b4-a0ac-c8db-4776-4482310251f2",
					"a9915312-f816-9936-226e-2cf453338b57",
				},
				Level: "raid1",
			},
			"bb994e65-dc69-2390-1ea5-09eed241ec96": {
				Type:   "filesystem",
				Device: "a2fc56c2-4788-320a-abef-17033ebdec41",
				FSType: "ext4",
				Mount:  "/",
			},
			"afa399ed-5bdb-5a40-8bb1-a9e181d2a85b": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 479,
					Type: "SSD SATA",
				},
			},
			"a9915312-f816-9936-226e-2cf453338b57": {
				Type:     "partition",
				Device:   "afa399ed-5bdb-5a40-8bb1-a9e181d2a85b",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"a2fc56c2-4788-320a-abef-17033ebdec41": {
				Type: "soft_raid",
				Members: []string{
					"1517b788-cd69-4cd7-17bf-fb1571fe5398",
					"2ceae753-dacd-ce55-2cb0-cd017ea30487",
				},
				Level: "raid1",
			},
			"695ea5b4-a0ac-c8db-4776-4482310251f2": {
				Type:     "partition",
				Device:   "258611b6-0647-5f3c-a2ec-b686355e58b5",
				Size:     1.0,
				Priority: ptr(0),
			},
			"2ceae753-dacd-ce55-2cb0-cd017ea30487": {
				Type:     "partition",
				Device:   "afa399ed-5bdb-5a40-8bb1-a9e181d2a85b",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"258611b6-0647-5f3c-a2ec-b686355e58b5": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 479,
					Type: "SSD SATA",
				},
			},
			"1517b788-cd69-4cd7-17bf-fb1571fe5398": {
				Type:     "partition",
				Device:   "258611b6-0647-5f3c-a2ec-b686355e58b5",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"134521be-a522-109b-6d05-23973eeb9a32": {
				Type:   "filesystem",
				Device: "d894a3f8-a17f-c141-76c4-e4ffdd240c56",
				FSType: "ext3",
				Mount:  "/boot",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 1)

		config := result[0]

		// Check soft raids - should be 1 since both RAIDs have same disk type and level
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1, "should have 1 RAID (both RAIDs have same disk type and level)")

		// Check that RAID name is readable
		sr := softRaids[0]
		name, ok := sr[dedicatedServerSchemaKeyName].(string)
		require.True(t, ok)
		assert.Equal(t, "new-raid1", name, "RAID name should be 'new-raid1'")
		assert.Equal(t, "raid1", sr[dedicatedServerSchemaKeyLevel], "RAID level should be 'raid1'")
		assert.Equal(t, "SSD SATA", sr[dedicatedServerSchemaKeyDiskType], "RAID disk type should be 'SSD SATA'")
		assert.Equal(t, 2, sr[dedicatedServerSchemaKeyDiskCount], "RAID should have 2 physical disks")

		// disk_config should be empty since all disks are in RAID
		diskConfigs, hasDiskConfigs := config[dedicatedServerSchemaKeyDiskConfig]
		assert.False(t, hasDiskConfigs || diskConfigs != nil, "disk_config should not be present when all disks are in RAID")

		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, diskPartitions, 2)

		var bootPartition, rootPartition map[string]interface{}
		for _, dp := range diskPartitions {
			if mount, ok := dp[dedicatedServerSchemaKeyMount].(string); ok {
				if mount == "/boot" {
					bootPartition = dp
				}
				if mount == "/" {
					rootPartition = dp
				}
			}
		}

		require.NotNil(t, bootPartition, "should have /boot partition")
		require.NotNil(t, rootPartition, "should have / partition")

		// Check sizes
		bootSize, ok := bootPartition[dedicatedServerSchemaKeySize].(float64)
		require.True(t, ok, "boot partition should have size")
		assert.Equal(t, float64(1), bootSize, "boot partition size should be 1")

		rootSize, ok := rootPartition[dedicatedServerSchemaKeySize].(float64)
		require.True(t, ok, "root partition should have size")
		assert.Equal(t, float64(-1), rootSize, "root partition size should be -1")

		// Check that partitions reference RAID by name
		bootRaid, ok := bootPartition[dedicatedServerSchemaKeyRaid].(string)
		require.True(t, ok, "boot partition should have raid")
		assert.Equal(t, "new-raid1", bootRaid, "boot partition should reference new-raid1")

		rootRaid, ok := rootPartition[dedicatedServerSchemaKeyRaid].(string)
		require.True(t, ok, "root partition should have raid")
		assert.Equal(t, "new-raid1", rootRaid, "root partition should reference new-raid1")
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		apiCfg := dedicated.PartitionsConfig{}
		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("SimpleDiskConfig", func(t *testing.T) {
		apiCfg := dedicated.PartitionsConfig{
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     1.0,
				Priority: ptr(0),
			},
			"fs1": {
				Type:   "filesystem",
				Device: "part1",
				FSType: "ext4",
				Mount:  "/",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 1)

		config := result[0]
		diskConfigs, ok := config[dedicatedServerSchemaKeyDiskConfig].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskConfigs, 1)

		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskPartitions, 1)
	})

	t.Run("MultipleRAIDTypes", func(t *testing.T) {
		// Test with 2 NVMe disks and 2 HDD disks - should create 2 separate RAID configs
		apiCfg := dedicated.PartitionsConfig{
			// NVMe drives
			"nvme-drive1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 479,
					Type: "SSD NVMe",
				},
			},
			"nvme-drive2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 479,
					Type: "SSD NVMe",
				},
			},
			// HDD drives
			"hdd-drive1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"hdd-drive2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// NVMe partitions (boot and root on different drives)
			"nvme-boot-part": {
				Type:     "partition",
				Device:   "nvme-drive1",
				Size:     1.0,
				Priority: ptr(0),
			},
			"nvme-root-part": {
				Type:     "partition",
				Device:   "nvme-drive2",
				Size:     -1.0,
				Priority: ptr(1),
			},
			// HDD partitions (backup)
			"hdd-backup-part1": {
				Type:     "partition",
				Device:   "hdd-drive1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"hdd-backup-part2": {
				Type:     "partition",
				Device:   "hdd-drive2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			// RAID arrays
			"nvme-boot-raid": {
				Type:    "soft_raid",
				Members: []string{"nvme-boot-part"},
				Level:   "raid1",
			},
			"nvme-root-raid": {
				Type:    "soft_raid",
				Members: []string{"nvme-root-part"},
				Level:   "raid1",
			},
			"hdd-backup-raid": {
				Type:    "soft_raid",
				Members: []string{"hdd-backup-part1", "hdd-backup-part2"},
				Level:   "raid1",
			},
			// Filesystems
			"boot-fs": {
				Type:   "filesystem",
				Device: "nvme-boot-raid",
				FSType: "ext4",
				Mount:  "/boot",
			},
			"root-fs": {
				Type:   "filesystem",
				Device: "nvme-root-raid",
				FSType: "ext4",
				Mount:  "/",
			},
			"backup-fs": {
				Type:   "filesystem",
				Device: "hdd-backup-raid",
				FSType: "ext4",
				Mount:  "/backup",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 1)

		config := result[0]

		// Should have 2 RAID configs: one for NVMe, one for HDD
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 2)

		// Check RAID names are different
		var nvmeRaid, hddRaid string
		var nvmeCount, hddCount int
		for _, sr := range softRaids {
			name := sr[dedicatedServerSchemaKeyName].(string)
			diskType := sr[dedicatedServerSchemaKeyDiskType].(string)
			count := sr[dedicatedServerSchemaKeyDiskCount].(int)
			switch diskType {
			case "SSD NVMe":
				nvmeRaid = name
				nvmeCount = count
			case "HDD SATA":
				hddRaid = name
				hddCount = count
			}
		}
		assert.NotEmpty(t, nvmeRaid, "should have NVMe RAID")
		assert.NotEmpty(t, hddRaid, "should have HDD RAID")
		assert.NotEqual(t, nvmeRaid, hddRaid, "RAID names should be different")
		assert.Equal(t, 2, nvmeCount, "NVMe RAID should have 2 disks")
		assert.Equal(t, 2, hddCount, "HDD RAID should have 2 disks")

		// Should have 3 partitions
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, diskPartitions, 3)
	})
}

func TestApiPartitionsConfigToSchema_Comprehensive(t *testing.T) {
	t.Run("SingleDiskNoRAID", func(t *testing.T) {
		// Single disk for OS, no RAID, other disks unconfigured
		apiCfg := dedicated.PartitionsConfig{
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 479,
					Type: "SSD SATA",
				},
			},
			"part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     1.0,
				Priority: ptr(0),
			},
			"fs1": {
				Type:   "filesystem",
				Device: "part1",
				FSType: "ext3",
				Mount:  "/boot",
			},
			"part2": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"fs2": {
				Type:   "filesystem",
				Device: "part2",
				FSType: "ext4",
				Mount:  "/",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 1)

		config := result[0]

		// No RAID configs (or empty array)
		softRaids, hasSoftRaids := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		assert.True(t, !hasSoftRaids || len(softRaids) == 0)

		// Has disk_config
		diskConfigs, ok := config[dedicatedServerSchemaKeyDiskConfig].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskConfigs, 1)

		// Has partitions
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskPartitions, 2)
	})

	t.Run("FourDisksSeparateNoRAID", func(t *testing.T) {
		// 4 HDD disks, each configured separately without RAID
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// Disk 1 partitions
			"part1-1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     1.0,
				Priority: ptr(0),
			},
			"fs1-1": {
				Type:   "filesystem",
				Device: "part1-1",
				FSType: "ext3",
				Mount:  "/boot",
			},
			"part1-2": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"fs1-2": {
				Type:   "filesystem",
				Device: "part1-2",
				FSType: "ext4",
				Mount:  "/",
			},
			// Disk 2 partitions
			"part2-1": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"fs2-1": {
				Type:   "filesystem",
				Device: "part2-1",
				FSType: "ext4",
				Mount:  "/data",
			},
			// Disk 3 partitions
			"part3-1": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"fs3-1": {
				Type:   "filesystem",
				Device: "part3-1",
				FSType: "xfs",
				Mount:  "/backup",
			},
			// Disk 4 partitions
			"part4-1": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"fs4-1": {
				Type:   "filesystem",
				Device: "part4-1",
				FSType: "ext4",
				Mount:  "/var",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// No RAID configs (or empty array)
		softRaids, hasSoftRaids := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		assert.True(t, !hasSoftRaids || len(softRaids) == 0)

		// Single disk_config (all same type)
		diskConfigs, ok := config[dedicatedServerSchemaKeyDiskConfig].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskConfigs, 1)
		assert.Equal(t, "disk-hdd-sata", diskConfigs[0][dedicatedServerSchemaKeyName])

		// Has partitions for all disks
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(diskPartitions), 4)
	})

	t.Run("FourDisksRAID10", func(t *testing.T) {
		// All 4 disks in RAID10
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID10 partitions (4 members)
			"raid10-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid10-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid10-part3": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid10-part4": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			// RAID10 array
			"raid10-array": {
				Type:    "soft_raid",
				Members: []string{"raid10-part1", "raid10-part2", "raid10-part3", "raid10-part4"},
				Level:   "raid10",
			},
			// Filesystem on RAID10
			"fs-root": {
				Type:   "filesystem",
				Device: "raid10-array",
				FSType: "ext4",
				Mount:  "/",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID10 config
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, "new-raid10", softRaids[0][dedicatedServerSchemaKeyName])
		assert.Equal(t, "raid10", softRaids[0][dedicatedServerSchemaKeyLevel])
		assert.Equal(t, "HDD SATA", softRaids[0][dedicatedServerSchemaKeyDiskType])
		assert.Equal(t, 4, softRaids[0][dedicatedServerSchemaKeyDiskCount])

		// No disk_config (all in RAID)
		_, hasDiskConfigs := config[dedicatedServerSchemaKeyDiskConfig]
		assert.False(t, hasDiskConfigs)

		// Has root partition
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskPartitions, 1)
		assert.Equal(t, "/", diskPartitions[0][dedicatedServerSchemaKeyMount])
		assert.Equal(t, "new-raid10", diskPartitions[0][dedicatedServerSchemaKeyRaid])
	})

	t.Run("FourDisksRAID0", func(t *testing.T) {
		// All 4 disks in RAID0
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID0 partitions
			"raid0-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-part3": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-part4": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			// RAID0 array
			"raid0-array": {
				Type:    "soft_raid",
				Members: []string{"raid0-part1", "raid0-part2", "raid0-part3", "raid0-part4"},
				Level:   "raid0",
			},
			// Filesystem on RAID0
			"fs-root": {
				Type:   "filesystem",
				Device: "raid0-array",
				FSType: "ext4",
				Mount:  "/",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID0 config with 4 disks
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, "new-raid0", softRaids[0][dedicatedServerSchemaKeyName])
		assert.Equal(t, 4, softRaids[0][dedicatedServerSchemaKeyDiskCount])

		// Root partition should have size -1 (not multiplied)
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, diskPartitions, 1)
		assert.Equal(t, float64(-1), diskPartitions[0][dedicatedServerSchemaKeySize])
	})

	t.Run("TwoRAID1", func(t *testing.T) {
		// 2 disks in RAID1, 2 disks in another RAID1
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID1 #1 (disk1, disk2)
			"raid1a-part": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1b-part": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1-array1": {
				Type:    "soft_raid",
				Members: []string{"raid1a-part", "raid1b-part"},
				Level:   "raid1",
			},
			"fs-root": {
				Type:   "filesystem",
				Device: "raid1-array1",
				FSType: "ext4",
				Mount:  "/",
			},
			// RAID1 #2 (disk3, disk4)
			"raid2a-part": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid2b-part": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid2-array": {
				Type:    "soft_raid",
				Members: []string{"raid2a-part", "raid2b-part"},
				Level:   "raid1",
			},
			"fs-backup": {
				Type:   "filesystem",
				Device: "raid2-array",
				FSType: "xfs",
				Mount:  "/backup",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID1 config (merged, 4 disks total)
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, "new-raid1", softRaids[0][dedicatedServerSchemaKeyName])
		assert.Equal(t, 4, softRaids[0][dedicatedServerSchemaKeyDiskCount])

		// Two partitions
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskPartitions, 2)
	})

	t.Run("TwoRAID0", func(t *testing.T) {
		// 2 disks in RAID0, 2 disks in another RAID0
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID0 #1 (disk1, disk2)
			"raid0a-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0a-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-array1": {
				Type:    "soft_raid",
				Members: []string{"raid0a-part1", "raid0a-part2"},
				Level:   "raid0",
			},
			"fs-root": {
				Type:   "filesystem",
				Device: "raid0-array1",
				FSType: "ext4",
				Mount:  "/",
			},
			// RAID0 #2 (disk3, disk4)
			"raid0b-part1": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0b-part2": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-array2": {
				Type:    "soft_raid",
				Members: []string{"raid0b-part1", "raid0b-part2"},
				Level:   "raid0",
			},
			"fs-data": {
				Type:   "filesystem",
				Device: "raid0-array2",
				FSType: "xfs",
				Mount:  "/data",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID0 config (merged, 4 disks total)
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, "new-raid0", softRaids[0][dedicatedServerSchemaKeyName])
		assert.Equal(t, 4, softRaids[0][dedicatedServerSchemaKeyDiskCount])
	})

	t.Run("RAID1PlusUnconfigured", func(t *testing.T) {
		// 2 disks in RAID1, 2 disks unconfigured
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID1 (disk1, disk2)
			"raid1-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     1.0,
				Priority: ptr(0),
			},
			"raid1-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     1.0,
				Priority: ptr(0),
			},
			"raid1-array": {
				Type:    "soft_raid",
				Members: []string{"raid1-part1", "raid1-part2"},
				Level:   "raid1",
			},
			"fs-boot": {
				Type:   "filesystem",
				Device: "raid1-array",
				FSType: "ext3",
				Mount:  "/boot",
			},
			"raid1-part3": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"raid1-part4": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(1),
			},
			"raid1-array2": {
				Type:    "soft_raid",
				Members: []string{"raid1-part3", "raid1-part4"},
				Level:   "raid1",
			},
			"fs-root": {
				Type:   "filesystem",
				Device: "raid1-array2",
				FSType: "ext4",
				Mount:  "/",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID1 config (2 disks)
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, 2, softRaids[0][dedicatedServerSchemaKeyDiskCount])

		// disk_config for unconfigured disks (disk3, disk4)
		diskConfigs, ok := config[dedicatedServerSchemaKeyDiskConfig].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskConfigs, 1)
		assert.Equal(t, "disk-hdd-sata", diskConfigs[0][dedicatedServerSchemaKeyName])
	})

	t.Run("RAID1PlusOneConfigured", func(t *testing.T) {
		// 2 disks in RAID1, 1 disk configured, 1 unconfigured
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID1 (disk1, disk2)
			"raid1-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1-array": {
				Type:    "soft_raid",
				Members: []string{"raid1-part1", "raid1-part2"},
				Level:   "raid1",
			},
			"fs-root": {
				Type:   "filesystem",
				Device: "raid1-array",
				FSType: "ext4",
				Mount:  "/",
			},
			// disk3 configured separately
			"disk3-part": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"fs-data": {
				Type:   "filesystem",
				Device: "disk3-part",
				FSType: "xfs",
				Mount:  "/data",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Single RAID1 config (2 disks)
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 1)
		assert.Equal(t, 2, softRaids[0][dedicatedServerSchemaKeyDiskCount])

		// disk_config for disk3 and disk4 (same type, merged)
		diskConfigs, ok := config[dedicatedServerSchemaKeyDiskConfig].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskConfigs, 1)

		// Partitions for root and data
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(diskPartitions), 2)
	})

	t.Run("MixedRAID0AndRAID1", func(t *testing.T) {
		// 2 disks in RAID0, 2 disks in RAID1
		apiCfg := dedicated.PartitionsConfig{
			// 4 physical disks
			"disk1": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk2": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk3": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			"disk4": {
				Type: "local_drive",
				Match: &dedicated.PartitionConfigItemMatch{
					Size: 1000,
					Type: "HDD SATA",
				},
			},
			// RAID0 (disk1, disk2)
			"raid0-part1": {
				Type:     "partition",
				Device:   "disk1",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-part2": {
				Type:     "partition",
				Device:   "disk2",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid0-array": {
				Type:    "soft_raid",
				Members: []string{"raid0-part1", "raid0-part2"},
				Level:   "raid0",
			},
			"fs-root": {
				Type:   "filesystem",
				Device: "raid0-array",
				FSType: "ext4",
				Mount:  "/",
			},
			// RAID1 (disk3, disk4)
			"raid1-part1": {
				Type:     "partition",
				Device:   "disk3",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1-part2": {
				Type:     "partition",
				Device:   "disk4",
				Size:     -1.0,
				Priority: ptr(0),
			},
			"raid1-array": {
				Type:    "soft_raid",
				Members: []string{"raid1-part1", "raid1-part2"},
				Level:   "raid1",
			},
			"fs-backup": {
				Type:   "filesystem",
				Device: "raid1-array",
				FSType: "xfs",
				Mount:  "/backup",
			},
		}

		result, err := apiPartitionsConfigToSchema(apiCfg)
		require.NoError(t, err)
		require.NotNil(t, result)

		config := result[0]

		// Two RAID configs: raid0 and raid1
		softRaids, ok := config[dedicatedServerSchemaKeySoftRaidConfig].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, softRaids, 2)

		// Find RAID0 and RAID1
		var raid0Found, raid1Found bool
		for _, sr := range softRaids {
			level := sr[dedicatedServerSchemaKeyLevel].(string)
			switch level {
			case "raid0":
				raid0Found = true
				assert.Equal(t, 2, sr[dedicatedServerSchemaKeyDiskCount])
			case "raid1":
				raid1Found = true
				assert.Equal(t, 2, sr[dedicatedServerSchemaKeyDiskCount])
			}
		}
		assert.True(t, raid0Found, "RAID0 should be present")
		assert.True(t, raid1Found, "RAID1 should be present")

		// Two partitions
		diskPartitions, ok := config[dedicatedServerSchemaKeyDiskPartitions].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, diskPartitions, 2)
	})
}

func ptr(v int) *int {
	return &v
}
