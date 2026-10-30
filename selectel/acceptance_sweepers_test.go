package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/selectel/dbaas-go"
	dbaas_v2_common "github.com/selectel/dbaas-go/v2/common"
	filestorage "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/selectel/filestorage-go/pkg/v1/share"
	"github.com/selectel/filestorage-go/pkg/v1/sharenetwork"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/keypairs"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/projects"
	"github.com/selectel/images-go/pkg/v1/image"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/network"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/selectel/vpc-go/pkg/v2/router"
	"github.com/selectel/vpc-go/pkg/v2/securitygroup"
	"github.com/selectel/vpc-go/pkg/v2/subnet"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/dbaas"
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

// testSweepDeleteTimeout limits the wait for all deletes of one sweeper run,
// so that the sweep job keeps time for the other sweepers.
const testSweepDeleteTimeout = 20 * time.Minute

// testSweepPollInterval is the pause between the reads of a deleted object.
const testSweepPollInterval = 10 * time.Second

// testSweepDBaaSV1Region is the pool where the DBaaS v1 tests create their
// datastores, each test in a project of its own.
const testSweepDBaaSV1Region = "ru-3"

func init() {
	resource.AddTestSweepers("selectel_vpc_router_v2", &resource.Sweeper{
		Name: "selectel_vpc_router_v2",
		F:    testSweepVPCRouters,
	})
	resource.AddTestSweepers("selectel_dbaas_clickhouse_datastore_v2", &resource.Sweeper{
		Name: "selectel_dbaas_clickhouse_datastore_v2",
		F:    testSweepDBaaSClickhouseDatastores,
	})
	resource.AddTestSweepers("selectel_filestorage_share_v2", &resource.Sweeper{
		Name: "selectel_filestorage_share_v2",
		F:    testSweepFileStorageShares,
	})
	resource.AddTestSweepers("selectel_vpc_network_v2", &resource.Sweeper{
		Name: "selectel_vpc_network_v2",
		Dependencies: []string{
			"selectel_vpc_router_v2", "selectel_dbaas_clickhouse_datastore_v2", "selectel_filestorage_share_v2",
		},
		F: testSweepVPCNetworks,
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
	resource.AddTestSweepers("selectel_compute_image_v2", &resource.Sweeper{
		Name: "selectel_compute_image_v2",
		F:    testSweepComputeImages,
	})
	resource.AddTestSweepers("selectel_vpc_keypair_v2", &resource.Sweeper{
		Name: "selectel_vpc_keypair_v2",
		F:    testSweepVPCKeypairs,
	})
	resource.AddTestSweepers("selectel_vpc_project_v2", &resource.Sweeper{
		Name: "selectel_vpc_project_v2",
		F:    testSweepVPCProjects,
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

// testSweepOld reports whether an object is older than testSweepMinAge. The
// creation time is RFC 3339, or UTC without a zone as DBaaS returns it. An
// object of unknown age stays.
func testSweepOld(createdAt string) bool {
	created, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		created, err = time.Parse("2006-01-02T15:04:05.999999", createdAt)
	}

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

// testSweepDBaaSClickhouseDatastores waits until the datastores are deleted:
// their servers keep ports in the test subnets, and the network sweeper runs
// next. A datastore that is already DELETING gets no second delete.
func testSweepDBaaSClickhouseDatastores(region string) error {
	ctx, cancel := context.WithTimeout(context.Background(), testSweepDeleteTimeout)
	defer cancel()
	projectID, err := testSweepProjectID()
	if err != nil {
		return err
	}
	selvpcClient, err := testAccConfigFromEnv().GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return err
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(DBaaSv2, region)
	if err != nil {
		return err
	}
	client, err := dbaas.NewDBAASClientV2(selvpcClient.GetXAuthToken(), endpoint.URL)
	if err != nil {
		return err
	}
	datastores, err := client.ClickHouse.GetDatastoreList(ctx)
	if err != nil {
		return fmt.Errorf("list clickhouse datastores: %w", err)
	}

	var errs []error
	var deleted []string
	for _, ds := range datastores {
		if !testSweepName(ds.Name) || !testSweepOld(ds.CreatedAt) || ds.Status == dbaas_v2_common.DatastoreStatusDeleting {
			continue
		}
		log.Printf("[INFO] Sweeping clickhouse datastore %s (%s, %s) in %s", ds.ID, ds.Name, ds.Status, region)
		if err := client.ClickHouse.DeleteDatastore(ctx, ds.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete clickhouse datastore %s (%s): %w", ds.ID, ds.Name, err))

			continue
		}
		deleted = append(deleted, ds.ID)
	}
	for _, id := range deleted {
		if err := waiters.WaitForDBaaSV2DatastoreDeleted(ctx, client.ClickHouse, id, testSweepDeleteTimeout); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// testSweepFileStorageShares deletes the shares with their access rules and
// waits until the shares and their service share networks are gone: the share
// server keeps a port in the test subnet, and the network sweeper runs next.
// A share that is already deleting gets no second delete.
func testSweepFileStorageShares(region string) error {
	ctx, cancel := context.WithTimeout(context.Background(), testSweepDeleteTimeout)
	defer cancel()
	projectID, err := testSweepProjectID()
	if err != nil {
		return err
	}
	client, err := testAccConfigFromEnv().FileStorageClient(ctx, projectID, region)
	if err != nil {
		return err
	}
	shares, err := share.ListDetail(ctx, client, share.ListOpts{})
	if err != nil {
		return fmt.Errorf("list file storages: %w", err)
	}

	var errs []error
	var deleted []share.View
	for _, s := range shares {
		if !testSweepName(s.Name) || !testSweepOld(s.CreatedAt) {
			continue
		}
		log.Printf("[INFO] Sweeping file storage %s (%s, %s) in %s", s.ID, s.Name, s.Status, region)
		if s.Status != "deleting" {
			if _, err := share.Delete(ctx, client, s.ID); err != nil && !filestorage.IsKind(err, filestorage.KindNotFound) {
				errs = append(errs, fmt.Errorf("delete file storage %s (%s): %w", s.ID, s.Name, err))

				continue
			}
		}
		deleted = append(deleted, s)
	}
	for _, s := range deleted {
		if err := testSweepFileStorageShareDeleted(ctx, client, s); err != nil {
			errs = append(errs, fmt.Errorf("file storage %s (%s): %w", s.ID, s.Name, err))
		}
	}

	return errors.Join(errs...)
}

// testSweepFileStorageShareDeleted waits for the share and then for the share
// network, which the fork deletes after the share when the share metadata
// carries delete_empty_share_network and the network has no other subnets.
// A share that failed to delete stays in error_deleting until an administrator
// removes it.
func testSweepFileStorageShareDeleted(ctx context.Context, client *filestorage.Client, s share.View) error {
	networkID := ""
	if _, flagged := s.Metadata["delete_empty_share_network"]; flagged && s.ShareNetworkID != "" {
		// The network is read while the share still holds it.
		network, _, err := sharenetwork.Get(ctx, client, s.ShareNetworkID)
		switch {
		case filestorage.IsKind(err, filestorage.KindNotFound):
		case err != nil:
			return fmt.Errorf("read share network %s: %w", s.ShareNetworkID, err)
		case len(network.Subnets) > 1:
			log.Printf("[INFO] Share network %s has %d subnets, the service keeps it", network.ID, len(network.Subnets))
		default:
			networkID = network.ID
		}
	}
	for {
		current, _, err := share.Get(ctx, client, s.ID)
		if filestorage.IsKind(err, filestorage.KindNotFound) {
			break
		}
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		if current.Status == "error_deleting" {
			return errors.New("delete failed with status error_deleting")
		}
		if err := testSweepPause(ctx); err != nil {
			return fmt.Errorf("wait for the delete: %w", err)
		}
	}
	if networkID == "" {
		return nil
	}
	for {
		_, _, err := sharenetwork.Get(ctx, client, networkID)
		if filestorage.IsKind(err, filestorage.KindNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read share network %s: %w", networkID, err)
		}
		if err := testSweepPause(ctx); err != nil {
			return fmt.Errorf("wait for the delete of share network %s: %w", networkID, err)
		}
	}
}

func testSweepPause(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(testSweepPollInterval):
		return nil
	}
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

// testSweepComputeImages removes the protection first: a protected image
// cannot be deleted, and a failed test may leave one.
func testSweepComputeImages(region string) error {
	ctx := context.Background()
	projectID, err := testSweepProjectID()
	if err != nil {
		return err
	}
	client, err := testAccConfigFromEnv().ImagesClient(ctx, projectID, region)
	if err != nil {
		return err
	}
	all, err := image.List(ctx, client, image.ListOpts{Owner: projectID})
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	var errs []error
	for _, i := range all {
		if !testSweepName(i.Name) || time.Since(i.CreatedAt) <= testSweepMinAge {
			continue
		}
		log.Printf("[INFO] Sweeping image %s (%s) in %s", i.ID, i.Name, region)
		if i.Protected {
			unprotect := []image.PatchOp{{Op: image.OpReplace, Path: "/protected", Value: false}}
			if _, _, err := image.Update(ctx, client, i.ID, unprotect); err != nil {
				errs = append(errs, fmt.Errorf("unprotect image %s (%s): %w", i.ID, i.Name, err))
				continue
			}
		}
		if _, err := image.Delete(ctx, client, i.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete image %s (%s): %w", i.ID, i.Name, err))
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

// testSweepVPCProjects deletes the test projects of the account with their
// DBaaS v1 datastores in testSweepDBaaSV1Region, whatever regions -sweep
// lists. Keystone does not return the creation time of a project, so a
// project takes the age of its datastores, and an empty one has no age
// check: the acceptance resource group keeps runs from overlapping. A project
// whose datastores are young, already deleting or fail to delete stays.
func testSweepVPCProjects(_ string) error {
	ctx, cancel := context.WithTimeout(context.Background(), testSweepDeleteTimeout)
	defer cancel()
	client, err := testAccConfigFromEnv().GetSelVPCClient()
	if err != nil {
		return err
	}
	all, _, err := projects.List(client)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	// The datastores of all projects are deleted first and then awaited
	// together, so that the slowest delete bounds the wait.
	var errs []error
	var sweeps []testSweepProject
	for _, p := range all {
		if !testSweepName(p.Name) {
			continue
		}
		sweep, err := testSweepDBaaSV1Datastores(ctx, p)
		if err != nil {
			errs = append(errs, fmt.Errorf("project %s (%s) stays: %w", p.ID, p.Name, err))

			continue
		}
		sweeps = append(sweeps, sweep)
	}
	for _, sweep := range sweeps {
		if err := sweep.waitDatastoresDeleted(ctx); err != nil {
			errs = append(errs, fmt.Errorf("project %s (%s) stays: %w", sweep.project.ID, sweep.project.Name, err))

			continue
		}
		log.Printf("[INFO] Sweeping project %s (%s)", sweep.project.ID, sweep.project.Name)
		if _, err := projects.Delete(client, sweep.project.ID); err != nil {
			errs = append(errs, fmt.Errorf("delete project %s (%s): %w", sweep.project.ID, sweep.project.Name, err))
		}
	}

	return errors.Join(errs...)
}

// testSweepProject is a test project whose datastores are being deleted.
type testSweepProject struct {
	project    *projects.Project
	client     *dbaas.API
	datastores []string
}

// testSweepDBaaSV1Datastores deletes the datastores of a test project. It
// fails when the project keeps a datastore: one that is not an old test
// object, or one already PENDING_DELETE, which gets no second delete and no
// wait because the service may keep it there for days.
func testSweepDBaaSV1Datastores(ctx context.Context, p *projects.Project) (testSweepProject, error) {
	sweep := testSweepProject{project: p}
	selvpcClient, err := testAccConfigFromEnv().GetSelVPCClientWithProjectScope(p.ID)
	if err != nil {
		return sweep, err
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(DBaaS, testSweepDBaaSV1Region)
	if err != nil {
		return sweep, err
	}
	sweep.client, err = dbaas.NewDBAASClient(selvpcClient.GetXAuthToken(), endpoint.URL)
	if err != nil {
		return sweep, err
	}
	datastores, err := sweep.client.Datastores(ctx, nil)
	if err != nil {
		return sweep, fmt.Errorf("list datastores: %w", err)
	}

	var errs []error
	for _, ds := range datastores {
		switch {
		case !testSweepName(ds.Name) || !testSweepOld(ds.CreatedAt):
			errs = append(errs, fmt.Errorf("datastore %s (%s) is not an old test datastore", ds.ID, ds.Name))
		case ds.Status == dbaas.StatusPendingDelete:
			errs = append(errs, fmt.Errorf("datastore %s (%s) is %s since %s", ds.ID, ds.Name, ds.Status, ds.UpdatedAt))
		default:
			log.Printf("[INFO] Sweeping datastore %s (%s, %s) of project %s", ds.ID, ds.Name, ds.Status, p.ID)
			if err := sweep.client.DeleteDatastore(ctx, ds.ID); err != nil {
				errs = append(errs, fmt.Errorf("delete datastore %s (%s): %w", ds.ID, ds.Name, err))

				continue
			}
			sweep.datastores = append(sweep.datastores, ds.ID)
		}
	}

	return sweep, errors.Join(errs...)
}

func (s testSweepProject) waitDatastoresDeleted(ctx context.Context) error {
	for _, id := range s.datastores {
		wait := &resource.StateChangeConf{
			Pending:    []string{strconv.Itoa(http.StatusOK)},
			Target:     []string{strconv.Itoa(http.StatusNotFound)},
			Refresh:    waiters.DBaaSDatastoreV1DeleteStateRefreshFunc(ctx, s.client, id),
			Timeout:    testSweepDeleteTimeout,
			Delay:      10 * time.Second,
			MinTimeout: 20 * time.Second,
		}
		if _, err := wait.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("wait for deleted datastore %s: %w", id, err)
		}
	}

	return nil
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
