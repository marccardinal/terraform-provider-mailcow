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

func resourceTlsPolicyMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTlsPolicyMapCreate,
		ReadContext:   resourceTlsPolicyMapRead,
		UpdateContext: resourceTlsPolicyMapUpdate,
		DeleteContext: resourceTlsPolicyMapDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTlsPolicyMapImport,
		},

		Schema: map[string]*schema.Schema{
			"dest": {
				Type:        schema.TypeString,
				Description: "Target domain or email address for the TLS policy",
				Required:    true,
				ForceNew:    true,
			},
			"policy": {
				Type:        schema.TypeString,
				Description: "TLS policy: none, may, encrypt, dane, fingerprint, verify, or secure",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"none", "may", "encrypt", "dane", "'dane", "fingerprint", "verify", "secure",
				}, false),
			},
			"parameters": {
				Type:        schema.TypeString,
				Description: "Additional Postfix TLS parameters (key=value pairs separated by spaces)",
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this TLS policy map entry is active",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceTlsPolicyMapImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceTlsPolicyMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateTlsPolicyMapRequest()
	mailcowCreateRequest.Set("dest", d.Get("dest").(string))
	mailcowCreateRequest.Set("policy", d.Get("policy").(string))
	mailcowCreateRequest.Set("parameters", d.Get("parameters").(string))
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceTlsPolicyMapCreate", d.Get("dest").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := tlsPolicyMapFindId(ctx, c, d.Get("dest").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceTlsPolicyMapRead(ctx, d, m)
}

func resourceTlsPolicyMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	id := d.Id()

	request := c.client.Api.MailcowGetTlsPolicyMap(ctx, id)
	entry, err := readRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	if entry["id"] == nil {
		return diag.FromErr(errors.New("tls policy map entry not found: " + id))
	}

	if err := d.Set("dest", fmt.Sprint(entry["dest"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("policy", fmt.Sprint(entry["policy"])); err != nil {
		return diag.FromErr(err)
	}
	params := fmt.Sprint(entry["parameters"])
	if params == "<nil>" {
		params = ""
	}
	if err := d.Set("parameters", params); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", fmt.Sprint(entry["active"]) == "1"); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return diags
}

func resourceTlsPolicyMapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Only active can change without ForceNew; the Mailcow API has no edit endpoint,
	// so we delete and recreate to apply the active change.
	if err := resourceTlsPolicyMapDelete(ctx, d, m); err != nil {
		return err
	}
	return resourceTlsPolicyMapCreate(ctx, d, m)
}

func resourceTlsPolicyMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteTlsPolicyMapRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func tlsPolicyMapFindId(ctx context.Context, c *APIClient, dest string) (string, error) {
	request := c.client.Api.MailcowGetTlsPolicyMapAll(ctx)
	entries, err := readAllRequest(request)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if fmt.Sprint(entry["dest"]) == dest {
			return fmt.Sprintf("%.0f", entry["id"].(float64)), nil
		}
	}
	return "", errors.New("tls policy map entry not found after create: " + dest)
}
