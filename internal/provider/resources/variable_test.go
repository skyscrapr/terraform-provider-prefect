package resources_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccVariableResource(name string, varname, value string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}
resource "prefect_variable" "test" {
	workspace_id = prefect_workspace.workspace.id
	name = "%s"
	value = "%s"
}
	`, name, name, varname, value)
}

func fixtureAccVariableResourceWithTags(name string, varname, value string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}
resource "prefect_variable" "test" {
	workspace_id = prefect_workspace.workspace.id
	name = "%s"
	value = "%s"
	tags = ["foo", "bar"]
}
	`, name, name, varname, value)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_variable(t *testing.T) {
	resourceName := "prefect_variable.test"

	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	varName := strings.ReplaceAll(randomName, "-", "_")
	varName2 := strings.ReplaceAll(randomName2, "-", "_")

	randomValue := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomValue2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the variable resource
				Config: fixtureAccVariableResource(randomName, varName, randomValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", varName),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue),
				),
			},
			{
				// Check updating name + value of the variable resource
				Config: fixtureAccVariableResource(randomName, varName2, randomValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", varName2),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue2),
				),
			},
			{
				// Check adding tags
				Config: fixtureAccVariableResourceWithTags(randomName, varName2, randomValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", varName2),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue2),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "bar"),
				),
			},
		},
	})
}
