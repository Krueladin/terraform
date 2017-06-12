package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMCosmosDBName_validation(t *testing.T) {
	str := acctest.RandString(50)
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "ab",
			ErrCount: 1,
		},
		{
			Value:    "abc",
			ErrCount: 0,
		},
		{
			Value:    str,
			ErrCount: 0,
		},
		{
			Value:    str + "a",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAzureRmCosmosDBName(tc.Value, "azurerm_cosmos_db")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AzureRM CosmosDB Name to trigger a validation error for '%s'", tc.Value)
		}
	}
}

func TestAccAzureRMCosmosDB_standard_boundedStaleness(t *testing.T) {

	ri := acctest.RandInt()
	config := testAccAzureRMCosmosDB_standard_boundedStaleness(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCosmosDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCosmosDBExists("azurerm_cosmos_db.test"),
				),
			},
		},
	})
}

func TestAccAzureRMCosmosDB_standard_eventualConsistency(t *testing.T) {

	ri := acctest.RandInt()
	config := testAccAzureRMCosmosDB_standard_eventualConsistency(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCosmosDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCosmosDBExists("azurerm_cosmos_db.test"),
				),
			},
		},
	})
}

func TestAccAzureRMCosmosDB_standardGeoReplicated(t *testing.T) {

	ri := acctest.RandInt()
	config := testAccAzureRMCosmosDB_standardGeoReplicated(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCosmosDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCosmosDBExists("azurerm_cosmos_db.test"),
				),
			},
		},
	})
}

func testCheckAzureRMCosmosDBDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).documentDbClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_cosmos_db" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(resourceGroup, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("CosmosDB instance still exists:\n%#v", resp)
		}
	}

	return nil
}

func testCheckAzureRMCosmosDBExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for CosmosDB instance: %s", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).documentDbClient

		resp, err := conn.Get(resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on documentDbClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: CosmosDB instance %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testAccAzureRMCosmosDB_standard_boundedStaleness(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
    name = "acctestRG-%d"
    location = "West US"
}
resource "azurerm_cosmos_db" "test" {
  name                = "acctest-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  offer_type          = "Standard"

  consistency_policy {
    consistency_level       = "BoundedStaleness"
    max_interval_in_seconds = 10
    max_staleness_prefix    = 200
  }

  failover_policy {
    location = "${azurerm_resource_group.test.location}"
    priority = 0
  }
}
`, rInt, rInt)
}

func testAccAzureRMCosmosDB_standard_eventualConsistency(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
    name = "acctestRG-%d"
    location = "West US"
}
resource "azurerm_cosmos_db" "test" {
  name                = "acctest-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  offer_type          = "Standard"

  consistency_policy {
    consistency_level       = "Eventual"
    max_interval_in_seconds = 10
    max_staleness_prefix    = 200
  }

  failover_policy {
    location = "${azurerm_resource_group.test.location}"
    priority = 0
  }
}
`, rInt, rInt)
}

func testAccAzureRMCosmosDB_standardGeoReplicated(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
    name = "acctestRG-%d"
    location = "West US"
}
resource "azurerm_cosmos_db" "test" {
  name                = "acctest-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  offer_type          = "Standard"

  consistency_policy {
    consistency_level       = "Eventual"
    max_interval_in_seconds = 10
    max_staleness           = 200
  }

  failover_policy {
    location = "${azurerm_resource_group.test.location}"
    priority = 0
  }

  failover_policy {
    location = "West Europe"
    priority = 1
  }
}
`, rInt, rInt)
}
