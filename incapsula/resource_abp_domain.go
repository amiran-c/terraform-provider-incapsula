package incapsula

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

func resourceABPDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceABPDomainCreate,
		Read:   resourceABPDomainRead,
		Update: resourceABPDomainUpdate,
		Delete: resourceABPDomainDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyParts := strings.Split(d.Id(), "-")
				if len(keyParts) == 5 {
					return []*schema.ResourceData{d}, nil
				} else {
					return nil, fmt.Errorf("ABP domain ID %s not found, can't import", d.Id())
				}
			},
		},

		Schema: map[string]*schema.Schema{
			// Required Arguments
			"account_id": {
				Description: "Identifier of the account the policy belongs to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site_id": {
				Description: "Identifier of the site the policy belongs to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"challenge_ip_lookup_mode": {
				Description: "The lookup settings.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"analysis_ip_lookup_mode": {
				Description: "The lookup settings.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"criteria": {
				Description: "Matches one or more domain names.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cookiescope": {
				Description: "The Domain attribute of the Set-Cookie header that is set by the ABP JavaScript.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"captcha_settings": {
				Description: "Configures the type of CAPTCHA to use.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"log_region": {
				Description: "The region where logs are stored, available values: [apac, australia, eu, usa].",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"obfuscate_path": {
				Description: "The recommended path to use to load the ABP JS.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"mobile_api_obfuscate_path": {
				Description: "The path used by mobile SDKs to communicate with the analysis host.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cookie_mode": {
				Description: "The SameSite policy of ABP cookies, available values: [lax, legacy, none_secure, lax_and_none_secure, lax_and_legacy, legacy_and_none_secure].",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"no_js_injection_paths": {
				Description: "Prevent JS injection in paths listed.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"unmasked_headers": {
				Description: "Headers to not be masked by CloudWAF.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceABPDomainRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	ID := d.Id()
	log.Printf("[DEBUG] ABP read domain ID: %s", ID)

	abpDomainResponse, err := client.GetABPDomain(ID)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error reading ABP policy ID %s: %s", ID, err)
	} else {
		log.Printf("[DEBUG] ABP read policy, got policy: %v", abpDomainResponse)
	}
	d.SetId(abpDomainResponse.ID)
	d.Set("account_id", abpDomainResponse.AccountID)
	d.Set("site_id", abpDomainResponse.SiteID)
	d.Set("challenge_ip_lookup_mode", abpDomainResponse.ChallengeIPLookupMode)
	d.Set("analysis_ip_lookup_mode", abpDomainResponse.AnalysisIPLookupMode)
	d.Set("criteria", abpDomainResponse.Criteria)
	d.Set("cookiescope", abpDomainResponse.CookieScope)
	d.Set("captcha_settings", abpDomainResponse.CaptchaSettings)
	d.Set("log_region", abpDomainResponse.LogRegion)
	d.Set("obfuscate_path", abpDomainResponse.ObfuscatePath)
	d.Set("mobile_api_obfuscate_path", abpDomainResponse.MobileAPIObfuscatePath)
	d.Set("cookie_mode", abpDomainResponse.CookieMode)
	d.Set("no_js_injection_paths", abpDomainResponse.NoJSInjectionPaths)
	d.Set("unmasked_headers", abpDomainResponse.UnmaskedHeaders)


	return nil
}

func resourceABPDomainUpdate(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	domain := ABPUpdateDomain{}
	// fillABPPolicyData(d, &domain)

	log.Printf("[DEBUG] ABP update domain, created domain for request: %v", domain)

	client := m.(*Client)
	if _, err := client.updateABPDomain(ID, &domain); err != nil {
		return err
	}

	return resourceABPDomainRead(d, m)
}

func resourceABPDomainCreate(d *schema.ResourceData, m interface{}) error {
	domain := ABPCreateDomain{}
	// fillABPPolicyData(d, &domain)

	log.Printf("[DEBUG] ABP create domain, created domain for request: %v", domain)

	client := m.(*Client)
	if s, err := client.createABPDomain(&domain); err != nil {
		return err
	} else {
		d.SetId(s.ID)
	}

	return resourceABPDomainRead(d, m)
}

func resourceABPDomainDelete(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	client := m.(*Client)

	log.Printf("[DEBUG] ABP delete domain %s", ID)

	return client.deleteABPDomain(ID)
}
