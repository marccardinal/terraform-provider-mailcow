# Terraform Provider: Mailcow

A Terraform provider for managing [Mailcow](https://github.com/mailcow/mailcow-dockerized) configurations — domains, mailboxes, aliases, DKIM keys, relay hosts, BCC rules, and more.

This is a fork of [l-with/terraform-provider-mailcow](https://github.com/l-with/terraform-provider-mailcow) with additional resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider from source)

## Usage

```terraform
terraform {
  required_providers {
    mailcow = {
      source  = "marccardinal/mailcow"
      version = "~> 1.0"
    }
  }
}

provider "mailcow" {
  host_name = "mail.example.com"
  api_key   = var.mailcow_api_key
}
```

## Resources

| Resource | Description |
|---|---|
| `mailcow_alias` | Email alias |
| `mailcow_bcc` | BCC map rule (sender or recipient) |
| `mailcow_dkim` | DKIM key for a domain |
| `mailcow_domain` | Mailcow domain |
| `mailcow_domain_admin` | Domain administrator account |
| `mailcow_domain_alias` | Domain alias |
| `mailcow_domain_policy` | Per-domain antispam whitelist/blacklist policy |
| `mailcow_fwdhost` | Forwarding host (trusted relay) |
| `mailcow_identity_provider_keycloak` | Keycloak identity provider |
| `mailcow_mailbox` | Mailbox |
| `mailcow_oauth2_client` | OAuth2 client |
| `mailcow_recipient_map` | Recipient address rewriting rule |
| `mailcow_relayhost` | Outbound relay host with sender-dependent transport |
| `mailcow_resource` | CalDAV booking resource (location, group, or thing) |
| `mailcow_rl_domain` | Outbound rate limit for a domain |
| `mailcow_rl_mailbox` | Outbound rate limit for a mailbox |
| `mailcow_syncjob` | IMAP sync job |
| `mailcow_tls_policy_map` | Outgoing TLS policy map override |

## Data Sources

| Data Source | Description |
|---|---|
| `mailcow_dkim` | Read a DKIM key |
| `mailcow_domain` | Read a domain |
| `mailcow_mailbox` | Read a mailbox |

## Example

```terraform
resource "mailcow_domain" "example" {
  domain = "example.com"
}

resource "mailcow_dkim" "example" {
  domain = mailcow_domain.example.domain
  length = 2048
}

resource "mailcow_mailbox" "admin" {
  domain     = mailcow_domain.example.domain
  local_part = "admin"
  password   = var.admin_password
}

resource "mailcow_relayhost" "smtp2go" {
  hostname = "[mail.smtp2go.com]:2525"
  username = var.smtp2go_username
  password = var.smtp2go_password
  domains  = [mailcow_domain.example.domain]
}
```

## Building

```bash
go build -o terraform-provider-mailcow .
```

## Testing

Acceptance tests require a running Mailcow instance. Set the following environment variables before running:

```bash
export MAILCOW_HOST_NAME=mail.example.com
export MAILCOW_API_KEY=your-api-key
export MAILCOW_INSECURE=false  # set true to skip TLS verification

make testacc
```

Some tests require a Keycloak identity provider configured on the Mailcow instance:

```bash
curl -k "https://$MAILCOW_HOST_NAME/api/v1/edit/identity-provider" \
  -X POST \
  -H "X-API-Key: $MAILCOW_API_KEY" \
  -H 'Content-Type: application/json' \
  --data '{
    "items": ["identity-provider"],
    "attr": {
      "authsource": "keycloak",
      "server_url": "https://auth.example.com",
      "realm": "mailcow",
      "client_id": "mailcow_terraform",
      "client_secret": "example",
      "redirect_url": "https://mail.example.com",
      "version": "26.1.3"
    }
  }'
```

Without this, mailbox tests will fail on the `authsource` attribute check.

## Known Limitations

- **Tags**: The Mailcow API always appends tags rather than replacing them, so tag management is not currently supported. See [upstream issue](https://github.com/mailcow/mailcow-dockerized/issues/4681).
- **User ACLs**: No API endpoint exists to read user ACLs, preventing a `mailcow_user_acl` resource.
- **App passwords**: The add app-password endpoint requires a user-level API key; admin API keys receive `access_denied`.

## License

MIT — see [LICENSE](LICENSE).

Original work Copyright (c) 2022 l-with.
Additional contributions Copyright (c) 2026 Marc Vieira-Cardinal.
