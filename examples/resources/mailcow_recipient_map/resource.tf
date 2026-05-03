resource "mailcow_recipient_map" "redirect" {
  old_address = "old@example.com"
  new_address = "new@example.com"
  active      = true
}
