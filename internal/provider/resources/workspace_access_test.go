package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkspaceAccessResourceForBot(name string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "developer" {
	name = "Developer"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = prefect_workspace.workspace.id
	workspace_role_id = data.prefect_workspace_role.developer.id
}`, name, name, name)
}

func fixtureAccWorkspaceAccessResourceUpdateForBot(name string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "runner" {
	name = "Runner"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = prefect_workspace.workspace.id
	workspace_role_id = data.prefect_workspace_role.runner.id
}`, name, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_bot_workspace_access(t *testing.T) {
	accessResourceName := "prefect_workspace_access.bot_access"
	botResourceName := "prefect_service_account.bot"
	workspaceResourceName := "prefect_workspace.workspace"
	developerRoleDatsourceName := "data.prefect_workspace_role.developer"
	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"

	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResourceForBot(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", developerRoleDatsourceName, "id"),
				),
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForBot(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				),
			},
		},
	})
}
