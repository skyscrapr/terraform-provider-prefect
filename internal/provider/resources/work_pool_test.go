package resources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkPoolCreate(name string, poolType string, paused bool) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}
resource "prefect_work_pool" "test" {
	name = "%s"
	type = "%s"
	workspace_id = prefect_workspace.workspace.id
	paused = %t
}
`, name, name, name, poolType, paused)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_pool(t *testing.T) {
	resourceName := "prefect_work_pool.test"
	workspaceResourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	poolType := "kubernetes"
	poolType2 := "ecs"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the work pool resource
				Config: fixtureAccWorkPoolCreate(randomName, poolType, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "type", poolType),
					resource.TestCheckResourceAttr(resourceName, "paused", "true"),
				),
			},
			{
				// Check that changing the paused state will update the resource in place
				Config: fixtureAccWorkPoolCreate(randomName, poolType, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "type", poolType),
					resource.TestCheckResourceAttr(resourceName, "paused", "false"),
				),
			},
			{
				// Check that changing the name will re-create the resource
				Config: fixtureAccWorkPoolCreate(randomName2, poolType, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "type", poolType),
					resource.TestCheckResourceAttr(resourceName, "paused", "false"),
				),
			},
			{
				// Check that changing the poolType will re-create the resource
				Config: fixtureAccWorkPoolCreate(randomName2, poolType2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "type", poolType2),
					resource.TestCheckResourceAttr(resourceName, "paused", "false"),
				),
			},
			// Import State checks - import by workspace_id,name (dynamic)
			{
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateIdFunc: getWorkPoolImportStateID(resourceName, workspaceResourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func getWorkPoolImportStateID(workPoolResourceName string, workspaceDatsourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceDatsource, exists := state.RootModule().Resources[workspaceDatsourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workspaceDatsourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceDatsource.Primary.ID)

		workPoolResource, exists := state.RootModule().Resources[workPoolResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workPoolResourceName)
		}
		workPoolName := workPoolResource.Primary.Attributes["name"]

		return fmt.Sprintf("%s,%s", workspaceID, workPoolName), nil
	}
}
