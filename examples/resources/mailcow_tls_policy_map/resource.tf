resource "mailcow_tls_policy_map" "enforce" {
  dest   = "example.com"
  policy = "encrypt"
  active = true
}

resource "mailcow_tls_policy_map" "dane" {
  dest       = "secure.example.com"
  policy     = "dane"
  parameters = "match=nexthop"
  active     = true
}
