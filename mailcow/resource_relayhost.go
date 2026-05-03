package mailcow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/l-with/terraform-provider-mailcow/api"
)

func resourceRelayhost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRelayhostCreate,
		ReadContext:   resourceRelayhostRead,
		UpdateContext: resourceRelayhostUpdate,
		DeleteContext: resourceRelayhostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRelayhostImport,
		},

		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Description: "Relay host and optional port, e.g. \"[mail.smtp2go.com]:2525\"",
				Required:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "Username for SMTP authentication",
				Optional:    true,
				Default:     "",
			},
			"password": {
				Type:        schema.TypeString,
				Description: "Password for SMTP authentication",
				Optional:    true,
				Sensitive:   true,
				Default:     "",
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this relayhost is active",
				Optional:    true,
				Default:     true,
			},
			"domains": {
				Type:        schema.TypeSet,
				Description: "Domains that use this relayhost for sender-dependent transport",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"mailboxes": {
				Type:        schema.TypeSet,
				Description: "Mailboxes that use this relayhost for sender-dependent transport",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceRelayhostImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func relayhostSetToStringSlice(d *schema.ResourceData, key string) []string {
	raw := d.Get(key).(*schema.Set).List()
	result := make([]string, len(raw))
	for i, v := range raw {
		result[i] = v.(string)
	}
	return result
}

func resourceRelayhostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateRelayhostRequest()
	mailcowCreateRequest.Set("hostname", d.Get("hostname").(string))
	mailcowCreateRequest.Set("username", d.Get("username").(string))
	mailcowCreateRequest.Set("password", d.Get("password").(string))
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRelayhostCreate", d.Get("hostname").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := response.GetRelayhostId()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*id)

	// Set domain/mailbox associations via a follow-up edit
	if d.Get("domains").(*schema.Set).Len() > 0 || d.Get("mailboxes").(*schema.Set).Len() > 0 {
		return resourceRelayhostUpdate(ctx, d, m)
	}

	return resourceRelayhostRead(ctx, d, m)
}

func resourceRelayhostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	id := d.Id()

	request := c.client.Api.MailcowGetRelayhost(ctx, id)
	relayhost, err := readRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	if relayhost["id"] == nil {
		return diag.FromErr(errors.New("relayhost not found: " + id))
	}

	if err := d.Set("hostname", fmt.Sprint(relayhost["hostname"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", fmt.Sprint(relayhost["username"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", relayhost["active"].(float64) == 1); err != nil {
		return diag.FromErr(err)
	}
	// password is not returned by the API (only password_short); keep state as-is

	// Parse comma-separated used_by_domains / used_by_mailboxes
	if err := d.Set("domains", splitCSV(fmt.Sprint(relayhost["used_by_domains"]))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mailboxes", splitCSV(fmt.Sprint(relayhost["used_by_mailboxes"]))); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return diags
}

func resourceRelayhostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateRelayhostRequest()

	if d.HasChange("hostname") {
		mailcowUpdateRequest.SetAttr("hostname", d.Get("hostname").(string))
	}
	if d.HasChange("username") {
		mailcowUpdateRequest.SetAttr("username", d.Get("username").(string))
	}
	if d.HasChange("password") {
		mailcowUpdateRequest.SetAttr("password", d.Get("password").(string))
	}
	if d.HasChange("active") {
		mailcowUpdateRequest.SetAttr("active", d.Get("active").(bool))
	}
	if d.HasChange("domains") {
		mailcowUpdateRequest.SetAttr("domains", relayhostSetToStringSlice(d, "domains"))
	}
	if d.HasChange("mailboxes") {
		mailcowUpdateRequest.SetAttr("mailboxes", relayhostSetToStringSlice(d, "mailboxes"))
	}

	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRelayhostUpdate", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRelayhostRead(ctx, d, m)
}

func resourceRelayhostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteRelayhostRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func splitCSV(s string) []string {
	if s == "" || s == "<nil>" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
