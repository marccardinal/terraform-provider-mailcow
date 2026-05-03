resource "mailcow_fwdhost" "trusted_relay" {
  hostname    = "relay.example.com"
  filter_spam = false
}
