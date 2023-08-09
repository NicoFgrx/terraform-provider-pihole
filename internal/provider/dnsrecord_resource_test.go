package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSResourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "pihole_dnsrecord" "test" {
  domain = "test.example.com"
  ip = "1.2.3.4"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify static values
					resource.TestCheckResourceAttr("pihole_dnsrecord.test", "domain", "test.example.com"),
					resource.TestCheckResourceAttr("pihole_dnsrecord.test", "ip", "1.2.3.4"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("pihole_dnsrecord.test", "id"),
					resource.TestCheckResourceAttrSet("pihole_dnsrecord.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pihole_dnsrecord.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "pihole_dnsrecord" "test" {
	domain = "test.example.com"
	ip = "2.3.4.5"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify static values
					resource.TestCheckResourceAttr("pihole_dnsrecord.test", "domain", "test.example.com"),
					resource.TestCheckResourceAttr("pihole_dnsrecord.test", "ip", "2.3.4.5"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("pihole_dnsrecord.test", "id"),
					resource.TestCheckResourceAttrSet("pihole_dnsrecord.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
