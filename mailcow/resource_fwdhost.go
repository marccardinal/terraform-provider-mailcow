package mailcow

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/l-with/terraform-provider-mailcow/api"
)

func resourceFwdhost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFwdhostCreate,
		ReadContext:   resourceFwdhostRead,
		DeleteContext: resourceFwdhostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceFwdhostImport,
		},

		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Description: "Hostname or IP of the forwarding host",
				Required:    true,
				ForceNew:    true,
			},
			"filter_spam": {
				Type:        schema.TypeBool,
				Description: "Whether to apply spam filtering to emails from this host",
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
		},
	}
}

func resourceFwdhostImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceFwdhostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	filterSpam := 0
	if d.Get("filter_spam").(bool) {
		filterSpam = 1
	}

	mailcowCreateRequest := api.NewCreateFwdhostRequest()
	mailcowCreateRequest.Set("hostname", d.Get("hostname").(string))
	mailcowCreateRequest.Set("filter_spam", filterSpam)

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceFwdhostCreate", d.Get("hostname").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("hostname").(string))
	return resourceFwdhostRead(ctx, d, m)
}

func resourceFwdhostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	hostname := d.Id()

	request := c.client.Api.MailcowGetFwdhostAll(ctx)
	hosts, err := readAllRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var found map[string]interface{}
	for _, h := range hosts {
		if fmt.Sprint(h["source"]) == hostname {
			found = h
			break
		}
	}
	if found == nil {
		return diag.FromErr(errors.New("forwarding host not found: " + hostname))
	}

	if err := d.Set("hostname", fmt.Sprint(found["source"])); err != nil {
		return diag.FromErr(err)
	}
	// keep_spam is "yes"/"no"; filter_spam is the inverse
	if err := d.Set("filter_spam", fmt.Sprint(found["keep_spam"]) != "yes"); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(hostname)
	return diags
}

func resourceFwdhostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	hostname := d.Id()

	// Collect all IPs for this hostname from get-all
	request := c.client.Api.MailcowGetFwdhostAll(ctx)
	hosts, err := readAllRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var ips []string
	for _, h := range hosts {
		if fmt.Sprint(h["source"]) == hostname {
			ips = append(ips, fmt.Sprint(h["host"]))
		}
	}
	if len(ips) == 0 {
		d.SetId("")
		return nil
	}

	// Delete uses the first IP (Mailcow removes all IPs for the host together)
	mailcowDeleteRequest := api.NewDeleteFwdhostRequest()
	mailcowDeleteRequest.SetItem(ips[0])
	d.SetId(ips[0])

	response, err := api.MailcowDeleteExecute(ctx, c.client, mailcowDeleteRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceFwdhostDelete", hostname)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
