package mailcow

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/l-with/terraform-provider-mailcow/api"
)

func resourceRecipientMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRecipientMapCreate,
		ReadContext:   resourceRecipientMapRead,
		DeleteContext: resourceRecipientMapDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRecipientMapImport,
		},

		Schema: map[string]*schema.Schema{
			"old_address": {
				Type:        schema.TypeString,
				Description: "The address whose incoming mail should be rewritten",
				Required:    true,
				ForceNew:    true,
			},
			"new_address": {
				Type:        schema.TypeString,
				Description: "The address that should receive the rewritten mail",
				Required:    true,
				ForceNew:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this recipient map entry is active",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRecipientMapImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceRecipientMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateRecipientMapRequest()
	mailcowCreateRequest.Set("recipient_map_old", d.Get("old_address").(string))
	mailcowCreateRequest.Set("recipient_map_new", d.Get("new_address").(string))
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceRecipientMapCreate", d.Get("old_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := recipientMapFindId(ctx, c, d.Get("old_address").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceRecipientMapRead(ctx, d, m)
}

func resourceRecipientMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	id := d.Id()

	request := c.client.Api.MailcowGetRecipientMap(ctx, id)
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(entries) == 0 || entries[0]["id"] == nil {
		return diag.FromErr(errors.New("recipient map entry not found: " + id))
	}
	entry := entries[0]

	if err := d.Set("old_address", fmt.Sprint(entry["recipient_map_old"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("new_address", fmt.Sprint(entry["recipient_map_new"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", fmt.Sprint(entry["active"]) == "1"); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return diags
}

func resourceRecipientMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteRecipientMapRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func recipientMapFindId(ctx context.Context, c *APIClient, oldAddress string) (string, error) {
	request := c.client.Api.MailcowGetRecipientMap(ctx, "all")
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if fmt.Sprint(entry["recipient_map_old"]) == oldAddress {
			return fmt.Sprintf("%.0f", entry["id"].(float64)), nil
		}
	}
	return "", errors.New("recipient map entry not found after create: " + oldAddress)
}
