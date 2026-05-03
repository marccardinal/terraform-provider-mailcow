package mailcow

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/l-with/terraform-provider-mailcow/api"
)

func resourceDomainAdmin() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainAdminCreate,
		ReadContext:   resourceDomainAdminRead,
		UpdateContext: resourceDomainAdminUpdate,
		DeleteContext: resourceDomainAdminDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainAdminImport,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Description: "Domain admin username",
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "Domain admin password",
				Required:    true,
				Sensitive:   true,
			},
			"domains": {
				Type:        schema.TypeSet,
				Description: "Domains this admin has access to",
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this domain admin account is active",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceDomainAdminImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceDomainAdminCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	domains := domainAdminSetToStringSlice(d, "domains")
	domainsStr := ""
	if len(domains) > 0 {
		domainsStr = domains[0]
		for _, dom := range domains[1:] {
			domainsStr += "," + dom
		}
	}

	mailcowCreateRequest := api.NewCreateDomainAdminRequest()
	mailcowCreateRequest.Set("username", d.Get("username").(string))
	mailcowCreateRequest.Set("password", d.Get("password").(string))
	mailcowCreateRequest.Set("password2", d.Get("password").(string))
	mailcowCreateRequest.Set("domains", domainsStr)
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceDomainAdminCreate", d.Get("username").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("username").(string))
	return resourceDomainAdminRead(ctx, d, m)
}

func resourceDomainAdminRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	username := d.Id()

	request := c.client.Api.MailcowGetDomainAdminAll(ctx)
	admins, err := readAllRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var admin map[string]interface{}
	for _, a := range admins {
		if fmt.Sprint(a["username"]) == username {
			admin = a
			break
		}
	}
	if admin == nil {
		return diag.FromErr(errors.New("domain admin not found: " + username))
	}

	if err := d.Set("username", fmt.Sprint(admin["username"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", fmt.Sprint(admin["active"]) == "1"); err != nil {
		return diag.FromErr(err)
	}

	// selected_domains is a []interface{} of strings
	if rawDomains, ok := admin["selected_domains"].([]interface{}); ok {
		domains := make([]string, len(rawDomains))
		for i, v := range rawDomains {
			domains[i] = fmt.Sprint(v)
		}
		if err := d.Set("domains", domains); err != nil {
			return diag.FromErr(err)
		}
	}

	// password is not returned by API; keep state as-is
	d.SetId(username)
	return diags
}

func resourceDomainAdminUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowUpdateRequest := api.NewUpdateDomainAdminRequest()

	if d.HasChange("password") {
		mailcowUpdateRequest.SetAttr("password", d.Get("password").(string))
		mailcowUpdateRequest.SetAttr("password2", d.Get("password").(string))
	}
	if d.HasChange("domains") {
		mailcowUpdateRequest.SetAttr("domains", domainAdminSetToStringSlice(d, "domains"))
	}
	if d.HasChange("active") {
		mailcowUpdateRequest.SetAttr("active", d.Get("active").(bool))
	}

	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceDomainAdminUpdate", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDomainAdminRead(ctx, d, m)
}

func resourceDomainAdminDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteDomainAdminRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func domainAdminSetToStringSlice(d *schema.ResourceData, key string) []string {
	raw := d.Get(key).(*schema.Set).List()
	result := make([]string, len(raw))
	for i, v := range raw {
		result[i] = v.(string)
	}
	return result
}
