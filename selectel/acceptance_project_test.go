package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/selectel/go-selvpcclient/v5/selvpcclient/resell/v2/projects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAccExtraProject creates a project besides INFRA_PROJECT_ID, for example
// to receive a shared image. Call it before other fixtures: its cleanup then
// runs last and deletes the project after everything in it.
func testAccExtraProject(t *testing.T) string {
	t.Helper()
	projectID, err := testAccCreateProject()
	require.NoError(t, err)
	t.Logf("Created acceptance project %s", projectID)
	t.Cleanup(func() {
		assert.NoError(t, testAccDeleteProject(projectID), "delete acceptance project %s", projectID)
	})

	return projectID
}

func testAccCreateProject() (string, error) {
	client, err := testAccConfigFromEnv().GetSelVPCClient()
	if err != nil {
		return "", err
	}
	project, _, err := projects.Create(client, projects.CreateOpts{
		Name: acctest.RandomWithPrefix("tf-acc-project"),
	})
	if err != nil {
		return "", err
	}

	return project.ID, nil
}

func testAccDeleteProject(projectID string) error {
	client, err := testAccConfigFromEnv().GetSelVPCClient()
	if err != nil {
		return err
	}
	_, err = projects.Delete(client, projectID)

	return err
}
