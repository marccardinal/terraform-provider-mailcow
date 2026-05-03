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

func resourceBcc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBccCreate,
		ReadContext:   resourceBccRead,
		UpdateContext: resourceBccUpdate,
		DeleteContext: resourceBccDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceBccImport,
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Description:  "BCC map type: \"sender\" (outbound) or \"recipient\" (inbound)",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"sender", "recipient"}, false),
			},
			"local_dest": {
				Type:        schema.TypeString,
				Description: "Local address or domain to match (e.g. \"user@example.com\" or \"@example.com\")",
				Required:    true,
				ForceNew:    true,
			},
			"bcc_dest": {
				Type:        schema.TypeString,
				Description: "Destination address to BCC",
				Required:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this BCC rule is active",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceBccImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceBccCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateBccRequest()
	mailcowCreateRequest.Set("type", d.Get("type").(string))
	mailcowCreateRequest.Set("local_dest", d.Get("local_dest").(string))
	mailcowCreateRequest.Set("bcc_dest", d.Get("bcc_dest").(string))
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceBccCreate", d.Get("local_dest").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// The create response does not return the new ID, so fetch all and match.
	id, err := bccFindId(ctx, c, d.Get("local_dest").(string), d.Get("type").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceBccRead(ctx, d, m)
}

func resourceBccRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	id := d.Id()

	request := c.client.Api.MailcowGetBcc(ctx, id)
	bcc, err := readRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	if bcc["id"] == nil {
		return diag.FromErr(errors.New("bcc rule not found: " + id))
	}

	if err := d.Set("type", fmt.Sprint(bcc["type"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_dest", fmt.Sprint(bcc["local_dest"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bcc_dest", fmt.Sprint(bcc["bcc_dest"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", bcc["active"].(float64) == 1); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return diags
}

func resourceBccUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateBccRequest()

	if d.HasChange("bcc_dest") {
		mailcowUpdateRequest.SetAttr("bcc_dest", d.Get("bcc_dest").(string))
	}
	if d.HasChange("active") {
		mailcowUpdateRequest.SetAttr("active", d.Get("active").(bool))
	}

	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceBccUpdate", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBccRead(ctx, d, m)
}

func resourceBccDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteBccRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func bccFindId(ctx context.Context, c *APIClient, localDest, bccType string) (string, error) {
	request := c.client.Api.MailcowGetBccAll(ctx)
	rules, err := readAllRequest(request)
	if err != nil {
		return "", err
	}
	for _, rule := range rules {
		if fmt.Sprint(rule["local_dest"]) == localDest && fmt.Sprint(rule["type"]) == bccType {
			return fmt.Sprintf("%.0f", rule["id"].(float64)), nil
		}
	}
	return "", errors.New("bcc rule not found after create: " + localDest)
}
