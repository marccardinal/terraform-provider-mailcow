package mailcow

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/l-with/terraform-provider-mailcow/api"
)

// Mailcow's add/rsetting answers without the new id, and descriptions aren't
// unique: Create finds the id by comparing the rule ids before and after.
// Mailcow trims content on edit but not on add, and the admin UI stores CRLF
// line endings: content is compared and sent normalized (rspamdSettingContent).

func resourceRspamdSetting() *schema.Resource {
	return &schema.Resource{
		Description: "Rule of Mailcow's Rspamd settings map (System → Configuration → Rspamd settings map). " +
			"Mailcow passes `content` to Rspamd as the body of a settings rule without checking its syntax: " +
			"a rule with a syntax error stops Rspamd from loading the whole settings map, Mailcow's own rules included. " +
			"Check the rule on the server after applying, e.g. with `rspamc`.",

		CreateContext: resourceRspamdSettingCreate,
		ReadContext:   resourceRspamdSettingRead,
		UpdateContext: resourceRspamdSettingUpdate,
		DeleteContext: resourceRspamdSettingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRspamdSettingImport,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type: schema.TypeString,
				Description: "Short description shown in the Mailcow admin. " +
					"Required: Mailcow's edit endpoint keeps the stored value when it receives an empty one.",
				Required: true,
				ValidateFunc: func(v interface{}, k string) ([]string, []error) {
					if strings.TrimSpace(v.(string)) == "" {
						return nil, []error{fmt.Errorf("%q must not be empty", k)}
					}
					return nil, nil
				},
			},
			"content": {
				Type: schema.TypeString,
				Description: "Body of the Rspamd settings rule, e.g. `priority`, conditions (`rcpt`, `from`, `header`…) and `apply { … }`. " +
					"Compared with line endings and surrounding whitespace normalized.",
				Required: true,
				ValidateFunc: func(v interface{}, k string) ([]string, []error) {
					if rspamdSettingContent(v.(string)) == "" {
						return nil, []error{fmt.Errorf("%q must not be empty", k)}
					}
					return nil, nil
				},
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return rspamdSettingContent(old) == rspamdSettingContent(new)
				},
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether Rspamd applies the rule",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

var rspamdSettingIdPattern = regexp.MustCompile(`^\d+$`)

func resourceRspamdSettingImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if !rspamdSettingIdPattern.MatchString(d.Id()) {
		return nil, fmt.Errorf("import id must be the numeric id of the rule (shown as \"ID #n\" in the Mailcow admin), got %q", d.Id())
	}
	return []*schema.ResourceData{d}, nil
}

func resourceRspamdSettingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	description := d.Get("description").(string)
	content := rspamdSettingContent(d.Get("content").(string))

	before, err := readAllRequest(c.client.Api.MailcowGetRspamdSettingAll(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	mailcowCreateRequest := api.NewCreateRspamdSettingRequest()
	mailcowCreateRequest.Set("desc", description)
	mailcowCreateRequest.Set("content", content)
	mailcowCreateRequest.Set("active", d.Get("active").(bool))

	request := c.client.Api.MailcowCreate(ctx).MailcowCreateRequest(*mailcowCreateRequest)
	response, _, err := c.client.Api.MailcowCreateExecute(request)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := checkResponse(response, "resourceRspamdSettingCreate", description); err != nil {
		return diag.FromErr(err)
	}

	after, err := readAllRequest(c.client.Api.MailcowGetRspamdSettingAll(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	id, err := rspamdSettingNewId(before, after, description, content)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	return resourceRspamdSettingRead(ctx, d, m)
}

func resourceRspamdSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	entry, err := readRequest(c.client.Api.MailcowGetRspamdSetting(ctx, d.Id()))
	if err != nil {
		return diag.FromErr(err)
	}
	if entry["id"] == nil {
		// Deleted outside Terraform (Mailcow answers {}): plan a new one.
		d.SetId("")
		return nil
	}

	if err := d.Set("description", fmt.Sprint(entry["desc"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content", fmt.Sprint(entry["content"])); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", fmt.Sprint(entry["active"]) == "1"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRspamdSettingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)

	// Send every field: edit keeps the stored value for any empty one.
	mailcowUpdateRequest := api.NewUpdateRspamdSettingRequest()
	mailcowUpdateRequest.SetAttr("desc", d.Get("description").(string))
	mailcowUpdateRequest.SetAttr("content", rspamdSettingContent(d.Get("content").(string)))
	mailcowUpdateRequest.SetAttr("active", d.Get("active").(bool))
	mailcowUpdateRequest.SetItem(d.Id())

	response, err := api.MailcowUpdateExecute(ctx, c.client, mailcowUpdateRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := checkResponse(response, "resourceRspamdSettingUpdate", d.Id()); err != nil {
		return diag.FromErr(err)
	}
	return resourceRspamdSettingRead(ctx, d, m)
}

func resourceRspamdSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*APIClient)
	diags, _ := mailcowDelete(ctx, d, api.NewDeleteRspamdSettingRequest(), c)
	return diags
}

// rspamdSettingContent normalizes rule content: LF line endings, no
// surrounding whitespace.
func rspamdSettingContent(content string) string {
	return strings.TrimSpace(strings.ReplaceAll(content, "\r\n", "\n"))
}

// rspamdSettingNewId returns the id of the rule that appears in after but not
// in before, with the given description and content. Anything other than
// exactly one match is an error rather than a guess.
func rspamdSettingNewId(before, after []map[string]interface{}, description, content string) (string, error) {
	existing := make(map[string]bool, len(before))
	for _, entry := range before {
		existing[rspamdSettingId(entry["id"])] = true
	}
	var matches []string
	for _, entry := range after {
		id := rspamdSettingId(entry["id"])
		if existing[id] {
			continue
		}
		if fmt.Sprint(entry["desc"]) == description && rspamdSettingContent(fmt.Sprint(entry["content"])) == content {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", errors.New("rspamd setting not found after create: " + description)
	default:
		return "", fmt.Errorf("rspamd setting %q: %d new rules match, can't tell which one was created (ids %s)",
			description, len(matches), strings.Join(matches, ", "))
	}
}

// rspamdSettingId formats a rule id, which the API returns as a JSON number
// or string depending on the database driver.
func rspamdSettingId(v interface{}) string {
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprint(v)
}
