resource "mailcow_domain_policy" "block_spam_domain" {
  domain      = "example.com"
  object_from = "*@spammers.example"
  object_list = "bl"
}

resource "mailcow_domain_policy" "allow_partner" {
  domain      = "example.com"
  object_from = "*@trusted-partner.example"
  object_list = "wl"
}
