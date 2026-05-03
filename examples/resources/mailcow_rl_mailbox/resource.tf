resource "mailcow_rl_mailbox" "example" {
  mailbox  = "user@example.com"
  rl_value = "10"
  rl_frame = "h"
}
