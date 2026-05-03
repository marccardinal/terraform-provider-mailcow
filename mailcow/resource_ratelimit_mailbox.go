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

func resourceRatelimitMailbox() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRatelimitMailboxCreate,
		ReadContext:   resourceRatelimitMailboxRead,
		UpdateContext: resourceRatelimitMailboxUpdate,
		DeleteContext: resourceRatelimitMailboxDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRatelimitMailboxImport,
		},

		Schema: map[string]*schema.Schema{
			"mailbox": {
				Type:        schema.TypeString,
				Description: "Mailbox address to apply the rate limit to",
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

func resourceRatelimitMailboxImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceRatelimitMailboxCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailbox := d.Get("mailbox").(string)

	mailcowUpdateRequest := api.NewUpdateRlMboxRequest()
	mailcowUpdateRequest.SetAttr("rl_value", d.Get("rl_value").(string))
	mailcowUpdateRequest.SetAttr("rl_frame", d.Get("rl_frame").(string))
	mailcowUpdateRequest.SetItem(mailbox)

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRatelimitMailboxCreate", mailbox)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mailbox)
	return resourceRatelimitMailboxRead(ctx, d, m)
}

func resourceRatelimitMailboxRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	mailbox := d.Id()

	request := c.client.Api.MailcowGetRlMbox(ctx, mailbox)
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var found map[string]interface{}
	for _, e := range entries {
		if fmt.Sprint(e["mailbox"]) == mailbox {
			found = e
			break
		}
	}
	if found == nil {
		return diag.FromErr(errors.New("mailbox ratelimit not found: " + mailbox))
	}

	if err := d.Set("mailbox", mailbox); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rl_value", fmt.Sprint(found["value"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rl_frame", fmt.Sprint(found["frame"])); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mailbox)
	return diags
}

func resourceRatelimitMailboxUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateRlMboxRequest()
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
	err = checkResponse(response, "resourceRatelimitMailboxUpdate", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRatelimitMailboxRead(ctx, d, m)
}

func resourceRatelimitMailboxDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateRlMboxRequest()
	mailcowUpdateRequest.SetAttr("rl_value", "")
	mailcowUpdateRequest.SetAttr("rl_frame", "")
	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRatelimitMailboxDelete", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
