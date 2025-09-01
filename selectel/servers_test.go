package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/servers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	serverslocal "github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
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

func newTestServersAPIClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) *serverslocal.ServiceClient {
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

	return serverslocal.NewClientV2(selvpcClient.GetXAuthToken(), url)
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
	localDrives := serverslocal.LocalDrives{
		"drive-ssd-1": {
			Type: "drive",
			Match: &serverslocal.LocalDriveMatch{
				Size: 1000,
				Type: "SSD",
			},
		},
		"drive-ssd-2": {
			Type: "drive",
			Match: &serverslocal.LocalDriveMatch{
				Size: 1000,
				Type: "SSD",
			},
		},
		"drive-hdd-1": {
			Type: "drive",
			Match: &serverslocal.LocalDriveMatch{
				Size: 2000,
				Type: "SATA",
			},
		},
	}

	defaultPartitions := []*serverslocal.PartitionConfigItem{
		{Mount: "/boot", Size: 1, FSType: "ext3"},
		{Mount: "/", Size: -1, FSType: "ext4"},
		{Mount: "swap", Size: 12, FSType: "swap"},
	}

	findItemsByType := func(cfg serverslocal.PartitionsConfig, itemType string) []*serverslocal.PartitionConfigItem {
		var items []*serverslocal.PartitionConfigItem
		for _, item := range cfg {
			if item.Type == itemType {
				items = append(items, item)
			}
		}

		return items
	}

	findFSByMount := func(cfg serverslocal.PartitionsConfig, mount string) *serverslocal.PartitionConfigItem {
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
		assert.Equal(t, "xfs", rootFS.FSType)

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
		assert.Contains(t, err.Error(), "can't find disk type for non-existent-raid")
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
		assert.Contains(t, err.Error(), "can't find disk for /")
	})
}
