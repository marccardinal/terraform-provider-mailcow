resource "mailcow_relayhost" "smtp2go" {
  hostname = "[mail.smtp2go.com]:2525"
  username = "myuser"
  password = "supersecret"
  active   = true

  domains = [
    "example.com",
    "example.org",
  ]
}
