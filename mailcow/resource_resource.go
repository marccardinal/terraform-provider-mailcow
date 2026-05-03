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

func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceCreate,
		ReadContext:   resourceResourceRead,
		DeleteContext: resourceResourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceResourceImport,
		},

		Schema: map[string]*schema.Schema{
			"local_part": {
				Type:        schema.TypeString,
				Description: "Local part of the resource email address (before @)",
				Required:    true,
				ForceNew:    true,
			},
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain for this resource",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the resource",
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"kind": {
				Type:         schema.TypeString,
				Description:  "Type of resource: location, group, or thing",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"location", "group", "thing"}, false),
			},
			"multiple_bookings": {
				Type:        schema.TypeInt,
				Description: "How many simultaneous bookings are allowed (-1 = unlimited, 1 = one at a time)",
				Optional:    true,
				Default:     0,
				ForceNew:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether this resource is active",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceResourceImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	mailcowCreateRequest := api.NewCreateResourceRequest()
	mailcowCreateRequest.Set("local_part", d.Get("local_part").(string))
	mailcowCreateRequest.Set("domain", d.Get("domain").(string))
	mailcowCreateRequest.Set("description", d.Get("description").(string))
	mailcowCreateRequest.Set("kind", d.Get("kind").(string))
	mailcowCreateRequest.Set("multiple_bookings", fmt.Sprintf("%d", d.Get("multiple_bookings").(int)))
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("local_part").(string) + "@" + d.Get("domain").(string)
	err = checkResponse(response, "resourceResourceCreate", name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)
	return resourceResourceRead(ctx, d, m)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*APIClient)
	name := d.Id()

	request := c.client.Api.MailcowGetResourceAll(ctx)
	resources, err := readAllRequest(request)
	if err != nil {
		return diag.FromErr(err)
	}

	var found map[string]interface{}
	for _, r := range resources {
		if fmt.Sprint(r["name"]) == name {
			found = r
			break
		}
	}
	if found == nil {
		return diag.FromErr(errors.New("resource not found: " + name))
	}

	if err := d.Set("local_part", fmt.Sprint(found["local_part"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain", fmt.Sprint(found["domain"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", fmt.Sprint(found["description"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kind", fmt.Sprint(found["kind"])); err != nil {
		return diag.FromErr(err)
	}
	if mb, ok := found["multiple_bookings"].(float64); ok {
		if err := d.Set("multiple_bookings", int(mb)); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("active", fmt.Sprint(found["active"]) == "1"); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)
	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	mailcowDeleteRequest := api.NewDeleteResourceRequest()
	diags, _ := mailcowDelete(ctx, d, mailcowDeleteRequest, c)
	return diags
}
