package mailcow

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMtaSts(t *testing.T) {
	domain := fmt.Sprintf("mta-sts-%s.domain-%s.xyz", randomLowerCaseString(4), randomLowerCaseString(4))
	policyId := regexp.MustCompile(`^\d{14}$`)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMtaSts(domain, "enforce", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "id", domain),
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "mode", "enforce"),
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "mx.0", "mail."+domain),
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "max_age", "604800"),
					resource.TestMatchResourceAttr("mailcow_mta_sts.mta_sts", "policy_id", policyId),
				),
			},
			{
				// The public policy page must read back exactly what was applied.
				Config:   testAccResourceMtaSts(domain, "enforce", "true"),
				PlanOnly: true,
			},
			{
				Config: testAccResourceMtaSts(domain, "testing", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "mode", "testing"),
					resource.TestMatchResourceAttr("mailcow_mta_sts.mta_sts", "policy_id", policyId),
				),
			},
			{
				Config: testAccResourceMtaSts(domain, "testing", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mailcow_mta_sts.mta_sts", "active", "false"),
				),
			},
			{
				Config:   testAccResourceMtaSts(domain, "testing", "false"),
				PlanOnly: true,
			},
		},
	})
}

func testAccResourceMtaSts(domain, mode, active string) string {
	return fmt.Sprintf(`
resource "mailcow_domain" "domain" {
  domain = "%[1]s"
}

resource "mailcow_mta_sts" "mta_sts" {
  domain = mailcow_domain.domain.domain
  mode   = "%[2]s"
  mx     = ["mail.%[1]s"]
  active = %[3]s
}
`, domain, mode, active)
}
