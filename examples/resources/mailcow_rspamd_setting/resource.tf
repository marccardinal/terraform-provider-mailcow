# Don't count a Reply-To on another domain as spoofing for the site's own
# form notifications (they reply to the visitor) — and only for them.
resource "mailcow_rspamd_setting" "form_notifications" {
  description = "example.com form notifications: allow visitor Reply-To"
  content     = <<-EOT
    priority = high;
    rcpt = "hello@example.com";
    header = {
      "From" = "/noreply@example\.com/i";
    }
    apply {
      symbols_disabled = ["SPOOF_REPLYTO"];
    }
  EOT
}
