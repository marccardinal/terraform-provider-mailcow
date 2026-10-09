resource "mailcow_mta_sts" "example" {
  domain  = "example.com"
  mode    = "enforce"
  mx      = ["mail.example.com"]
  max_age = 604800
}

# Publish the policy id in DNS, e.g. with the Cloudflare provider
resource "cloudflare_dns_record" "mta_sts" {
  zone_id = var.zone_id
  name    = "_mta-sts"
  type    = "TXT"
  content = "v=STSv1; id=${mailcow_mta_sts.example.policy_id};"
  ttl     = 1
}
