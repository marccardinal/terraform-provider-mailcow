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

func resourceDomainPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainPolicyCreate,
		ReadContext:   resourceDomainPolicyRead,
		DeleteContext: resourceDomainPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain to which this policy applies",
				Required:    true,
				ForceNew:    true,
			},
			"object_from": {
				Type:        schema.TypeString,
				Description: "Exact address or wildcard pattern to match (e.g. \"*@baddomain.tld\")",
				Required:    true,
				ForceNew:    true,
			},
			"object_list": {
				Type:         schema.TypeString,
				Description:  "Policy list type: \"wl\" (whitelist) or \"bl\" (blacklist)",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"wl", "bl"}, false),
			},
		},
	}
}

func resourceDomainPolicyImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceDomainPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateDomainPolicyRequest()
	mailcowCreateRequest.Set("domain", d.Get("domain").(string))
	mailcowCreateRequest.Set("object_from", d.Get("object_from").(string))
	mailcowCreateRequest.Set("object_list", d.Get("object_list").(string))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	err = checkResponse(response, "resourceDomainPolicyCreate", d.Get("domain").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := domainPolicyFindId(ctx, c, d.Get("domain").(string), d.Get("object_list").(string), d.Get("object_from").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceDomainPolicyRead(ctx, d, m)
}

func resourceDomainPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	id := d.Id()
	domain := d.Get("domain").(string)
	objectList := d.Get("object_list").(string)

	entry, err := domainPolicyGetById(ctx, c, domain, objectList, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if entry == nil {
		return diag.FromErr(errors.New("domain policy not found: " + id))
	}

	if err := d.Set("domain", fmt.Sprint(entry["object"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object_from", fmt.Sprint(entry["value"])); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return diags
}

func resourceDomainPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteDomainPolicyRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}

func domainPolicyFindId(ctx context.Context, c *APIClient, domain, objectList, objectFrom string) (string, error) {
	endpoint := "/api/v1/get/policy_bl_domain/{id}"
	if objectList == "wl" {
		endpoint = "/api/v1/get/policy_wl_domain/{id}"
	}
	request := c.client.Api.MailcowGetDomainPolicy(ctx, domain, endpoint)
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if fmt.Sprint(entry["value"]) == objectFrom {
			return fmt.Sprintf("%.0f", entry["prefid"].(float64)), nil
		}
	}
	return "", errors.New("domain policy not found after create: " + objectFrom)
}

func domainPolicyGetById(ctx context.Context, c *APIClient, domain, objectList, id string) (map[string]interface{}, error) {
	endpoint := "/api/v1/get/policy_bl_domain/{id}"
	if objectList == "wl" {
		endpoint = "/api/v1/get/policy_wl_domain/{id}"
	}
	request := c.client.Api.MailcowGetDomainPolicy(ctx, domain, endpoint)
	entries, err := readListFromGetRequest(request)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if fmt.Sprintf("%.0f", entry["prefid"].(float64)) == id {
			return entry, nil
		}
	}
	return nil, nil
}
