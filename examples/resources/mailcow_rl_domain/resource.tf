resource "mailcow_rl_domain" "example" {
  domain   = "example.com"
  rl_value = "100"
  rl_frame = "h"
}
