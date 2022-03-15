// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: You have opted to include helpful guiding comments. These comments are
// meant to teach and remind. However, they should be removed before submitting
// your work in a PR. Thank you!

package scaffold

// TIP: This is a common set of imports but not fully customized to your code
// since your code hasn't been written yet. Make sure you, your IDE, or
// goimports -w <file> fixes these imports. 
import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scaffold"
	"github.com/aws/aws-sdk-go-v2/service/scaffold/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCheese() *schema.Resource {
	// In the schema, list each of the arguments and attributes in snake case
	// (e.g., delete_automated_backups).
	// 
	// Arguments can be assigned a value in the configuration while attributes
	// can be read as output. You will typically find arguments in the Input
	// struct for the create operation.
	Use "Computed: true," when you need to read information
	// from AWS or detect drift. "ValidateFunc" is helpful to catch errors
	// before anything is sent to AWS. With long-running configurations
	// especially, this is very helpful.
	return &schema.Resource{
		CreateWithoutTimeout: resourceCheeseCreate,
		ReadWithoutTimeout:   resourceCheeseRead,
		UpdateWithoutTimeout: resourceCheeseUpdate,
		DeleteWithoutTimeout: resourceCheeseDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"abuse_contact_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"abuse_contact_phone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_contact": contactSchema,
			"admin_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"auto_renew": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_server": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"glue_ips": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 2,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"registrant_contact": contactSchema,
			"registrant_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"registrar_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registrar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reseller": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":         tftags.TagsSchema(),
			"tags_all":     tftags.TagsSchemaComputed(),
			"tech_contact": contactSchema,
			"tech_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transfer_lock": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"whois_server": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCheeseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	domainName := d.Get("domain_name").(string)
	domainDetail, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return diag.Errorf("error reading Route 53 Domains Domain (%s): %s", domainName, err)
	}

	d.SetId(aws.ToString(domainDetail.DomainName))

	var adminContact, registrantContact, techContact *types.ContactDetail

	if v, ok := d.GetOk("admin_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.AdminContact) {
			adminContact = v
		}
	}

	if v, ok := d.GetOk("registrant_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.RegistrantContact) {
			registrantContact = v
		}
	}

	if v, ok := d.GetOk("tech_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.TechContact) {
			techContact = v
		}
	}

	if adminContact != nil || registrantContact != nil || techContact != nil {
		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, registrantContact, techContact, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if adminPrivacy, registrantPrivacy, techPrivacy := d.Get("admin_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool); adminPrivacy != aws.ToBool(domainDetail.AdminPrivacy) || registrantPrivacy != aws.ToBool(domainDetail.RegistrantPrivacy) || techPrivacy != aws.ToBool(domainDetail.TechPrivacy) {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), adminPrivacy, registrantPrivacy, techPrivacy, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if v := d.Get("auto_renew").(bool); v != aws.ToBool(domainDetail.AutoRenew) {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), v); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
		nameservers := expandNameservers(v.([]interface{}))

		if !reflect.DeepEqual(nameservers, domainDetail.Nameservers) {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), nameservers, d.Timeout(schema.TimeoutCreate)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if v := d.Get("transfer_lock").(bool); v != hasDomainTransferLock(domainDetail.StatusList) {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{}))).IgnoreConfig(ignoreTagsConfig)
	oldTags := tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := UpdateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return diag.Errorf("error updating Route 53 Domains Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCheeseRead(ctx, d, meta)
}

func resourceCheeseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainDetail, err := findDomainDetailByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Domains Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	d.Set("abuse_contact_email", domainDetail.AbuseContactEmail)
	d.Set("abuse_contact_phone", domainDetail.AbuseContactPhone)
	if domainDetail.AdminContact != nil {
		if err := d.Set("admin_contact", []interface{}{flattenContactDetail(domainDetail.AdminContact)}); err != nil {
			return diag.Errorf("error setting admin_contact: %s", err)
		}
	} else {
		d.Set("admin_contact", nil)
	}
	d.Set("admin_privacy", domainDetail.AdminPrivacy)
	d.Set("auto_renew", domainDetail.AutoRenew)
	if domainDetail.CreationDate != nil {
		d.Set("creation_date", aws.ToTime(domainDetail.CreationDate).Format(time.RFC3339))
	} else {
		d.Set("creation_date", nil)
	}
	d.Set("domain_name", domainDetail.DomainName)
	if domainDetail.ExpirationDate != nil {
		d.Set("expiration_date", aws.ToTime(domainDetail.ExpirationDate).Format(time.RFC3339))
	} else {
		d.Set("expiration_date", nil)
	}
	if err := d.Set("name_server", flattenNameservers(domainDetail.Nameservers)); err != nil {
		return diag.Errorf("error setting name_servers: %s", err)
	}
	if domainDetail.RegistrantContact != nil {
		if err := d.Set("registrant_contact", []interface{}{flattenContactDetail(domainDetail.RegistrantContact)}); err != nil {
			return diag.Errorf("error setting registrant_contact: %s", err)
		}
	} else {
		d.Set("registrant_contact", nil)
	}
	d.Set("registrant_privacy", domainDetail.RegistrantPrivacy)
	d.Set("registrar_name", domainDetail.RegistrarName)
	d.Set("registrar_url", domainDetail.RegistrarUrl)
	d.Set("reseller", domainDetail.Reseller)
	statusList := domainDetail.StatusList
	d.Set("status_list", statusList)
	if domainDetail.TechContact != nil {
		if err := d.Set("tech_contact", []interface{}{flattenContactDetail(domainDetail.TechContact)}); err != nil {
			return diag.Errorf("error setting tech_contact: %s", err)
		}
	} else {
		d.Set("tech_contact", nil)
	}
	d.Set("tech_privacy", domainDetail.TechPrivacy)
	d.Set("transfer_lock", hasDomainTransferLock(statusList))
	if domainDetail.UpdatedDate != nil {
		d.Set("updated_date", aws.ToTime(domainDetail.UpdatedDate).Format(time.RFC3339))
	} else {
		d.Set("updated_date", nil)
	}
	d.Set("whois_server", domainDetail.WhoIsServer)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceCheeseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	if d.HasChanges("admin_contact", "registrant_contact", "tech_contact") {
		var adminContact, registrantContact, techContact *types.ContactDetail

		if key := "admin_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				adminContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if key := "registrant_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				registrantContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if key := "tech_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				techContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, registrantContact, techContact, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("admin_privacy", "registrant_privacy", "tech_privacy") {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), d.Get("admin_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("auto_renew") {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), d.Get("auto_renew").(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name_server") {
		if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), expandNameservers(v.([]interface{})), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("transfer_lock") {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), d.Get("transfer_lock").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating Route 53 Domains Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCheeseRead(ctx, d, meta)
}
