package mailcow

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/l-with/terraform-provider-mailcow/api"
)

func resourceRatelimitDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRatelimitDomainCreate,
		ReadContext:   resourceRatelimitDomainRead,
		UpdateContext: resourceRatelimitDomainUpdate,
		DeleteContext: resourceRatelimitDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRatelimitDomainImport,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain to apply the rate limit to",
				Required:    true,
				ForceNew:    true,
			},
			"rl_value": {
				Type:        schema.TypeString,
				Description: "Rate limit value (e.g. \"10\", \"50\")",
				Required:    true,
			},
			"rl_frame": {
				Type:         schema.TypeString,
				Description:  "Rate limit time frame: s (second), m (minute), h (hour)",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"s", "m", "h"}, false),
			},
		},
	}
}

func resourceRatelimitDomainImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceRatelimitDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	domain := d.Get("domain").(string)

	mailcowUpdateRequest := api.NewUpdateRlDomainRequest()
	mailcowUpdateRequest.SetAttr("rl_value", d.Get("rl_value").(string))
	mailcowUpdateRequest.SetAttr("rl_frame", d.Get("rl_frame").(string))
	mailcowUpdateRequest.SetItem(domain)

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRatelimitDomainCreate", domain)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain)
	return resourceRatelimitDomainRead(ctx, d, m)
}

func resourceRatelimitDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	domain := d.Id()

	request := c.client.Api.MailcowGetRlDomain(ctx, domain)
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var found map[string]interface{}
	for _, e := range entries {
		if fmt.Sprint(e["domain"]) == domain {
			found = e
			break
		}
	}
	if found == nil {
		return diag.FromErr(errors.New("domain ratelimit not found: " + domain))
	}

	if err := d.Set("domain", domain); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rl_value", fmt.Sprint(found["value"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rl_frame", fmt.Sprint(found["frame"])); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain)
	return diags
}

func resourceRatelimitDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateRlDomainRequest()
	if d.HasChange("rl_value") {
		mailcowUpdateRequest.SetAttr("rl_value", d.Get("rl_value").(string))
	}
	if d.HasChange("rl_frame") {
		mailcowUpdateRequest.SetAttr("rl_frame", d.Get("rl_frame").(string))
	}
	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRatelimitDomainUpdate", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRatelimitDomainRead(ctx, d, m)
}

func resourceRatelimitDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateRlDomainRequest()
	mailcowUpdateRequest.SetAttr("rl_value", "")
	mailcowUpdateRequest.SetAttr("rl_frame", "")
	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRatelimitDomainDelete", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
