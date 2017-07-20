package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMDnsAAAARecord_basic(t *testing.T) {
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccAzureRMDnsAAAARecord_basic, ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMDnsAaaaRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDnsAaaaRecordExists("azurerm_dns_aaaa_record.test"),
				),
			},
		},
	})
}

func TestAccAzureRMDnsAAAARecord_updateRecords(t *testing.T) {
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAzureRMDnsAAAARecord_basic, ri, ri, ri)
	postConfig := fmt.Sprintf(testAccAzureRMDnsAAAARecord_updateRecords, ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMDnsAaaaRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDnsAaaaRecordExists("azurerm_dns_aaaa_record.test"),
					resource.TestCheckResourceAttr(
						"azurerm_dns_aaaa_record.test", "records.#", "2"),
				),
			},

			resource.TestStep{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDnsAaaaRecordExists("azurerm_dns_aaaa_record.test"),
					resource.TestCheckResourceAttr(
						"azurerm_dns_aaaa_record.test", "records.#", "3"),
				),
			},
		},
	})
}

func TestAccAzureRMDnsAAAARecord_withTags(t *testing.T) {
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAzureRMDnsAAAARecord_withTags, ri, ri, ri)
	postConfig := fmt.Sprintf(testAccAzureRMDnsAAAARecord_withTagsUpdate, ri, ri, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMDnsAaaaRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDnsAaaaRecordExists("azurerm_dns_aaaa_record.test"),
					resource.TestCheckResourceAttr(
						"azurerm_dns_aaaa_record.test", "tags.%", "2"),
				),
			},

			resource.TestStep{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDnsAaaaRecordExists("azurerm_dns_aaaa_record.test"),
					resource.TestCheckResourceAttr(
						"azurerm_dns_aaaa_record.test", "tags.%", "1"),
				),
			},
		},
	})
}

func testCheckAzureRMDnsAaaaRecordExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		aaaaName := rs.Primary.Attributes["name"]
		zoneName := rs.Primary.Attributes["zone_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for DNS AAAA record: %s", aaaaName)
		}

		conn := testAccProvider.Meta().(*ArmClient).dnsClient
		resp, err := conn.Get(resourceGroup, zoneName, aaaaName, dns.AAAA)
		if err != nil {
			return fmt.Errorf("Bad: Get AAAA RecordSet: %v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: DNS AAAA record %s (resource group: %s) does not exist", aaaaName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMDnsAaaaRecordDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).dnsClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_dns_aaaa_record" {
			continue
		}

		aaaaName := rs.Primary.Attributes["name"]
		zoneName := rs.Primary.Attributes["zone_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(resourceGroup, zoneName, aaaaName, dns.AAAA)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("DNS AAAA record still exists:\n%#v", resp.RecordSetProperties)
		}

	}

	return nil
}

var testAccAzureRMDnsAAAARecord_basic = `
resource "azurerm_resource_group" "test" {
    name = "acctestRG_%d"
    location = "West US"
}
resource "azurerm_dns_zone" "test" {
    name = "acctestzone%d.com"
    resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dns_aaaa_record" "test" {
    name = "myarecord%d"
    resource_group_name = "${azurerm_resource_group.test.name}"
    zone_name = "${azurerm_dns_zone.test.name}"
    ttl = "300"
    records = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]
}
`

var testAccAzureRMDnsAAAARecord_updateRecords = `
resource "azurerm_resource_group" "test" {
    name = "acctestRG_%d"
    location = "West US"
}
resource "azurerm_dns_zone" "test" {
    name = "acctestzone%d.com"
    resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dns_aaaa_record" "test" {
    name = "myarecord%d"
    resource_group_name = "${azurerm_resource_group.test.name}"
    zone_name = "${azurerm_dns_zone.test.name}"
    ttl = "300"
    records = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006", "::1"]
}
`

var testAccAzureRMDnsAAAARecord_withTags = `
resource "azurerm_resource_group" "test" {
    name = "acctestRG_%d"
    location = "West US"
}
resource "azurerm_dns_zone" "test" {
    name = "acctestzone%d.com"
    resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dns_aaaa_record" "test" {
    name = "myarecord%d"
    resource_group_name = "${azurerm_resource_group.test.name}"
    zone_name = "${azurerm_dns_zone.test.name}"
    ttl = "300"
    records = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]

    tags {
	environment = "Production"
	cost_center = "MSFT"
    }
}
`

var testAccAzureRMDnsAAAARecord_withTagsUpdate = `
resource "azurerm_resource_group" "test" {
    name = "acctestRG_%d"
    location = "West US"
}
resource "azurerm_dns_zone" "test" {
    name = "acctestzone%d.com"
    resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dns_aaaa_record" "test" {
    name = "myarecord%d"
    resource_group_name = "${azurerm_resource_group.test.name}"
    zone_name = "${azurerm_dns_zone.test.name}"
    ttl = "300"
    records = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]

    tags {
	environment = "staging"
    }
}
`
