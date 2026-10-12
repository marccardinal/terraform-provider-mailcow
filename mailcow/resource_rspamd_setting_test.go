package mailcow

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The rule only matches a recipient at example.invalid, so it never applies to
// real mail on the Mailcow instance under test.
func TestAccResourceRspamdSetting(t *testing.T) {
	rcpt := fmt.Sprintf("tf-acc-%s@example.invalid", randomLowerCaseString(8))
	description := "tf-acc " + rcpt
	var id string

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRspamdSetting(description, rcpt, "SPOOF_REPLYTO", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("mailcow_rspamd_setting.test", "id", rspamdSettingIdPattern),
					resource.TestCheckResourceAttr("mailcow_rspamd_setting.test", "description", description),
					resource.TestCheckResourceAttr("mailcow_rspamd_setting.test", "active", "true"),
					func(s *terraform.State) error {
						id = s.RootModule().Resources["mailcow_rspamd_setting.test"].Primary.ID
						return nil
					},
				),
			},
			{
				// The heredoc's trailing newline must not show as a change.
				Config:   testAccResourceRspamdSetting(description, rcpt, "SPOOF_REPLYTO", "true"),
				PlanOnly: true,
			},
			{
				Config: testAccResourceRspamdSetting(description+" (edited)", rcpt, "MIME_MA_MISSING_HTML", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mailcow_rspamd_setting.test", "description", description+" (edited)"),
					resource.TestCheckResourceAttrWith("mailcow_rspamd_setting.test", "content", func(v string) error {
						if rspamdSettingContent(v) != rspamdSettingContent(testAccRspamdSettingContent(rcpt, "MIME_MA_MISSING_HTML")) {
							return fmt.Errorf("content not updated: %q", v)
						}
						return nil
					}),
					// Edited in place, not replaced.
					resource.TestCheckResourceAttrWith("mailcow_rspamd_setting.test", "id", func(v string) error {
						if v != id {
							return errors.New("id changed from " + id + " to " + v)
						}
						return nil
					}),
				),
			},
			{
				Config: testAccResourceRspamdSetting(description+" (edited)", rcpt, "MIME_MA_MISSING_HTML", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mailcow_rspamd_setting.test", "active", "false"),
				),
			},
			{
				Config:   testAccResourceRspamdSetting(description+" (edited)", rcpt, "MIME_MA_MISSING_HTML", "false"),
				PlanOnly: true,
			},
			{
				ResourceName:      "mailcow_rspamd_setting.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Read stores content as Mailcow returns it (no trailing newline).
				ImportStateVerifyIgnore: []string{"content"},
			},
		},
	})
}

func testAccRspamdSettingContent(rcpt, symbol string) string {
	return fmt.Sprintf(`priority = low;
rcpt = "%s";
apply {
  symbols_disabled = ["%s"];
}
`, rcpt, symbol)
}

func testAccResourceRspamdSetting(description, rcpt, symbol, active string) string {
	return fmt.Sprintf(`
resource "mailcow_rspamd_setting" "test" {
  description = %q
  content     = <<-EOT
%sEOT
  active      = %s
}
`, description, testAccRspamdSettingContent(rcpt, symbol), active)
}
