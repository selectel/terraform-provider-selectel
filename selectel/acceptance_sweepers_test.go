package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/keypairs"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
)

// The sweepers delete what failed acceptance runs leave behind:
//
//	go test ./selectel -sweep=ru-6,ru-9 -sweep-allow-failures -timeout 30m
//
// -sweep lists the regions of the INFRA_PROJECT_ID project to sweep. Keypairs
// and service users belong to the account and are swept in every region run.
// A sweeper takes only objects with a test name, and in the project only
// objects older than testSweepMinAge, so a run in progress keeps its objects.
const testSweepMinAge = time.Hour

func init() {
	resource.AddTestSweepers("selectel_vpc_router_v2", &resource.Sweeper{
		Name: "selectel_vpc_router_v2",
		F:    testSweepVPCRouters,
	})
	resource.AddTestSweepers("selectel_vpc_network_v2", &resource.Sweeper{
		Name:         "selectel_vpc_network_v2",
		Dependencies: []string{"selectel_vpc_router_v2"},
		F:            testSweepVPCNetworks,
	})
	resource.AddTestSweepers("selectel_vpc_secgroup_v2", &resource.Sweeper{
		Name:         "selectel_vpc_secgroup_v2",
		Dependencies: []string{"selectel_vpc_network_v2"},
		F:            testSweepVPCSecGroups,
	})
	resource.AddTestSweepers("selectel_compute_volume_v3", &resource.Sweeper{
		Name: "selectel_compute_volume_v3",
		F:    testSweepComputeVolumes,
	})
	resource.AddTestSweepers("selectel_vpc_keypair_v2", &resource.Sweeper{
		Name: "selectel_vpc_keypair_v2",
		F:    testSweepVPCKeypairs,
	})
	resource.AddTestSweepers("selectel_iam_serviceuser_v1", &resource.Sweeper{
		Name:         "selectel_iam_serviceuser_v1",
		Dependencies: []string{"selectel_vpc_keypair_v2"},
		F:            testSweepIAMServiceUsers,
	})
}

// testSweepName reports whether a test created the object. The acceptance
// account itself is named tf_acceptance_tests and must stay.
func testSweepName(name string) bool {
	if name == "tf_acceptance_tests" {
		return false
	}

	return strings.HasPrefix(name, "tf-acc") || strings.HasPrefix(name, "tf_acc")
}

// testSweepRouterInterface reports whether Neutron attached the port to a
// router: plain, distributed and HA routers use their own device owners.
func testSweepRouterInterface(deviceOwner string) bool {
	switch deviceOwner {
	case "network:router_interface",
		"network:router_interface_distributed",
		"network:ha_router_replicated_interface":
		return true
	}

	return false
}

// testSweepOld reports whether an object with an RFC 3339 creation time is
// older than testSweepMinAge. An object of unknown age stays.
func testSweepOld(createdAt string) bool {
	created, err := time.Parse(time.RFC3339, createdAt)

	return err == nil && time.Since(created) > testSweepMinAge
}

func testSweepProjectID() (string, error) {
	projectID := os.Getenv("INFRA_PROJECT_ID")
	if projectID == "" {
		return "", errors.New("INFRA_PROJECT_ID must be set for the sweepers")
	}

	return projectID, nil
}

func testSweepVPCClient(region string) (*vpc.Client, string, error) {
	projectID, err := testSweepProjectID()
	if err != nil {
		return nil, "", err
	}
	client, err := vpcV2Client(testAccConfigFromEnv(), projectID, region)

	return client, projectID, err
}

func testSweepProjectFilter(projectID string, filters map[string][]string) vpc.ListOptions {
	all := map[string][]string{"project_id": {projectID}}
	maps.Copy(all, filters)

	return vpc.ListOptions{SelectionOptions: vpc.SelectionOptions{Filters: all}}
}

// testSweepVPCRouters detaches the router interfaces before the delete,
// which Neutron refuses while a subnet is attached.
func testSweepVPCRouters(region string) error {
	ctx := context.Background()
	client, projectID, err := testSweepVPCClient(region)
	if err != nil {
		return err
	}
	routers, err := router.List(ctx, client, testSweepProjectFilter(projectID, nil))
	if err != nil {
		return fmt.Errorf("list routers: %w", err)
	}

	var errs []error
	for _, r := range routers {
		if !testSweepName(r.Name) || !testSweepOld(r.CreatedAt) {
			continue
		}
		ports, err := port.List(ctx, client, testSweepProjectFilter(projectID, map[string][]string{
			"device_id": {r.ID},
		}))
		if err != nil {
			errs = append(errs, fmt.Errorf("list ports of router %s: %w", r.ID, err))

			continue
		}
		for _, p := range ports {
			if testSweepRouterInterface(p.DeviceOwner) {
				if _, err := router.RemoveInterface(ctx, client, r.ID, router.InterfaceRequest{PortID: p.ID}); err != nil {
					errs = append(errs, fmt.Errorf("detach port %s from router %s: %w", p.ID, r.ID, err))
				}
			}
		}
		log.Printf("[INFO] Sweeping router %s (%s) in %s", r.ID, r.Name, region)
		if err := router.Delete(ctx, client, r.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete router %s (%s): %w", r.ID, r.Name, err))
		}
	}

	return errors.Join(errs...)
}

// testSweepVPCNetworks deletes a network with its subnets. A port that a
// server or another service owns keeps the network, and its delete fails.
func testSweepVPCNetworks(region string) error {
	ctx := context.Background()
	client, projectID, err := testSweepVPCClient(region)
	if err != nil {
		return err
	}
	networks, err := network.List(ctx, client, testSweepProjectFilter(projectID, nil))
	if err != nil {
		return fmt.Errorf("list networks: %w", err)
	}

	var errs []error
	for _, n := range networks {
		if !testSweepName(n.Name) || !testSweepOld(n.CreatedAt) {
			continue
		}
		log.Printf("[INFO] Sweeping network %s (%s) in %s", n.ID, n.Name, region)
		if err := testSweepVPCNetwork(ctx, client, projectID, n); err != nil {
			errs = append(errs, fmt.Errorf("network %s (%s): %w", n.ID, n.Name, err))
		}
	}

	return errors.Join(errs...)
}

func testSweepVPCNetwork(ctx context.Context, client *vpc.Client, projectID string, n network.Network) error {
	byNetwork := map[string][]string{"network_id": {n.ID}}
	ports, err := port.List(ctx, client, testSweepProjectFilter(projectID, byNetwork))
	if err != nil {
		return fmt.Errorf("list ports: %w", err)
	}
	for _, p := range ports {
		switch {
		case testSweepRouterInterface(p.DeviceOwner):
			_, err = router.RemoveInterface(ctx, client, p.DeviceID, router.InterfaceRequest{PortID: p.ID})
		case p.DeviceOwner == "":
			err = port.Delete(ctx, client, p.ID)
		default:
			// Neutron removes the DHCP ports with the network.
			continue
		}
		if err != nil {
			return fmt.Errorf("remove port %s: %w", p.ID, err)
		}
	}

	subnets, err := subnet.List(ctx, client, testSweepProjectFilter(projectID, byNetwork))
	if err != nil {
		return fmt.Errorf("list subnets: %w", err)
	}
	for _, s := range subnets {
		if err := subnet.Delete(ctx, client, s.ID); err != nil {
			return fmt.Errorf("delete subnet %s: %w", s.ID, err)
		}
	}

	return network.Delete(ctx, client, n.ID)
}

func testSweepVPCSecGroups(region string) error {
	ctx := context.Background()
	client, projectID, err := testSweepVPCClient(region)
	if err != nil {
		return err
	}
	groups, err := securitygroup.List(ctx, client, testSweepProjectFilter(projectID, nil))
	if err != nil {
		return fmt.Errorf("list security groups: %w", err)
	}

	var errs []error
	for _, g := range groups {
		if !testSweepName(g.Name) || !testSweepOld(g.CreatedAt) {
			continue
		}
		log.Printf("[INFO] Sweeping security group %s (%s) in %s", g.ID, g.Name, region)
		if err := securitygroup.Delete(ctx, client, g.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete security group %s (%s): %w", g.ID, g.Name, err))
		}
	}

	return errors.Join(errs...)
}

// testSweepComputeVolumes deletes the volumes with their snapshots. Attached
// volumes stay: their server is not a test object of this project.
func testSweepComputeVolumes(region string) error {
	projectID, err := testSweepProjectID()
	if err != nil {
		return err
	}
	client, err := testAccOpenStackServiceClient(projectID, region, "volumev3")
	if err != nil {
		return err
	}
	pages, err := volumes.List(client, volumes.ListOpts{}).AllPages()
	if err != nil {
		return fmt.Errorf("list volumes: %w", err)
	}
	all, err := volumes.ExtractVolumes(pages)
	if err != nil {
		return fmt.Errorf("list volumes: %w", err)
	}

	var errs []error
	for _, v := range all {
		if !testSweepName(v.Name) || time.Since(v.CreatedAt) <= testSweepMinAge || len(v.Attachments) > 0 {
			continue
		}
		log.Printf("[INFO] Sweeping volume %s (%s) in %s", v.ID, v.Name, region)
		if err := volumes.Delete(client, v.ID, volumes.DeleteOpts{Cascade: true}).ExtractErr(); err != nil {
			errs = append(errs, fmt.Errorf("delete volume %s (%s): %w", v.ID, v.Name, err))
		}
	}

	return errors.Join(errs...)
}

func testSweepVPCKeypairs(_ string) error {
	client, err := testAccConfigFromEnv().GetSelVPCClient()
	if err != nil {
		return err
	}
	all, _, err := keypairs.List(client)
	if err != nil {
		return fmt.Errorf("list keypairs: %w", err)
	}

	var errs []error
	for _, k := range all {
		if !testSweepName(k.Name) {
			continue
		}
		log.Printf("[INFO] Sweeping keypair %s of user %s", k.Name, k.UserID)
		if _, err := keypairs.Delete(client, k.Name, k.UserID); err != nil {
			errs = append(errs, fmt.Errorf("delete keypair %s of user %s: %w", k.Name, k.UserID, err))
		}
	}

	return errors.Join(errs...)
}

// testSweepIAMServiceUsers has no age check: IAM does not return the
// creation time. The acceptance resource group keeps runs from overlapping.
func testSweepIAMServiceUsers(_ string) error {
	ctx := context.Background()
	client, diags := getIAMClient(testAccConfigFromEnv())
	if diags.HasError() {
		return fmt.Errorf("iam client: %s", diags[0].Summary)
	}
	list, err := client.ServiceUsers.List(ctx)
	if err != nil {
		return fmt.Errorf("list service users: %w", err)
	}

	var errs []error
	for _, u := range list.Users {
		if !testSweepName(u.Name) {
			continue
		}
		log.Printf("[INFO] Sweeping service user %s (%s)", u.ID, u.Name)
		if err := client.ServiceUsers.Delete(ctx, u.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete service user %s (%s): %w", u.ID, u.Name, err))
		}
	}

	return errors.Join(errs...)
}
