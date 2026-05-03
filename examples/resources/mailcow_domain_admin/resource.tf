resource "mailcow_domain_admin" "support" {
  username = "support-admin"
  password = "supersecret"
  active   = true

  domains = [
    "example.com",
    "example.org",
  ]
}
