resource "mailcow_bcc" "outbound" {
  type       = "sender"
  local_dest = "@example.com"
  bcc_dest   = "archive@example.com"
  active     = true
}

resource "mailcow_bcc" "inbound" {
  type       = "recipient"
  local_dest = "user@example.com"
  bcc_dest   = "archive@example.com"
  active     = true
}
