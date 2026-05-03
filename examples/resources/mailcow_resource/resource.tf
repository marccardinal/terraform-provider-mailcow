resource "mailcow_resource" "boardroom" {
  local_part        = "boardroom"
  domain            = "example.com"
  description       = "Main boardroom"
  kind              = "location"
  multiple_bookings = -1
  active            = true
}
