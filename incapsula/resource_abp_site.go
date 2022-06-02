package incapsula

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"reflect"
	"strings"
)

func resourceABPSite() *schema.Resource {
	return &schema.Resource{
		Create: resourceABPSiteCreate,
		Read:   resourceABPSiteRead,
		Update: resourceABPSiteUpdate,
		Delete: resourceABPSiteDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyParts := strings.Split(d.Id(), "-")
				if len(keyParts) == 5 {
					return importResourcesByID(d, meta.(*Client))
				} else {
					return importResourcesByName(d, meta.(*Client))
				}
			},
		},

		Schema: map[string]*schema.Schema{
			// Required Arguments
			/*
				"id": {
					Description: "Numeric identifier of the site.",
					Type:        schema.TypeString,
					Computed:    true,
				},
			*/
			"account_id": {
				Description: "Numeric identifier of the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the account.",
				Type:        schema.TypeString,
				Required:    true,
				//Computed:    true,
			},
			"mx_hostname_id": {
				Description: "if site belongs to Imperva WAF Gateway, this will be set to the MX hostname.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"selectors": {
				Description: "List of selectors configured for the site.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"criteria_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"criteria_value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"analysis_rate_limiting": {
							Type:     schema.TypeString,
							Required: true,
						},
						"derived_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func importResourcesByID(d *schema.ResourceData, client *Client) ([]*schema.ResourceData, error) {
	if _, err := client.GetABPSite(d.Id()); err == nil {
		return []*schema.ResourceData{d}, nil
	}
	return nil, fmt.Errorf("ABP site ID %s not found, can't import", d.Id())
}

func importResourcesByName(d *schema.ResourceData, client *Client) ([]*schema.ResourceData, error) {
	// Save the site name from the ID to use later for finding the sites
	siteName := d.Id()
	found := false

	sites, err := client.GetABPSitesForAccountWithoutID()
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] ABP site import, fetched %d ABP sites: %v", len(sites), sites)

	// Create a list of resources and add by matching site name
	results := []*schema.ResourceData{d}
	for _, site := range sites {
		if siteName == site.Name {
			log.Printf("[DEBUG] ABP site import adding new site ID (%s) , name (%s)", site.ID, site.Name)
			var cd *schema.ResourceData
			if found == false {
				d.SetId(site.ID)
				found = true
				log.Printf("[DEBUG] ABP site import found wanted resource %v , results: %v", cd, results)
			} else {
				cd = resourceABPSite().Data(nil)
				cd.SetId(site.ID)
				cd.SetType("incapsula_abp_site")
				results = append(results, cd)
				log.Printf("[DEBUG] ABP site import created new resource %v , results: %v", cd, results)
			}
		} else {
			log.Printf("[DEBUG] ABP site import skipping current site name (%s) because it's not (%s)", site.Name, siteName)
		}
	}
	log.Printf("[DEBUG] ABP site import, importing %d resources: %v", len(results), results)

	return results, nil
}

func resourceABPSiteRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	ID := d.Id()

	keyParts := strings.Split(ID, "-")
	if len(keyParts) != 5 {
		d.SetId("")
		return fmt.Errorf("ABP read site, ID was not resolved on import, can't fetch site (%s)", ID)
	}
	log.Printf("[DEBUG] ABP read site ID: %s", ID)

	abpSiteResponse, err := client.GetABPSite(ID)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error reading ABP site ID %s: %s", ID, err)
	} else {
		log.Printf("[DEBUG] ABP read site, got site: %v", abpSiteResponse)
	}
	d.SetId(abpSiteResponse.ID)
	d.Set("name", abpSiteResponse.Name)
	d.Set("account_id", abpSiteResponse.AccountID)
	d.Set("mx_hostname_id", abpSiteResponse.MxHostnameID)

	if len(abpSiteResponse.Selectors) > 0 {
		if err := d.Set("selectors", getSelectorsList(abpSiteResponse.Selectors)); err != nil {
			return fmt.Errorf("error handling ABP site, error setting selectors: %s", err)
		}
	}

	return nil
}

func getSelectorsList(selectors []ABPSiteSelector) []map[string]interface{} {
	log.Printf("selectors: %v", selectors)
	if len(selectors) == 0 {
		return nil
	}
	ret := make([]map[string]interface{}, len(selectors))
	log.Printf("selectors: %v", ret)

	for i, selector := range selectors {
		log.Printf("selector index %d: %v", i, selector)
		var criteriaName string
		var criteriaValue string
		switch {
		case selector.Criteria.PathPrefix != nil:
			criteriaName = "path_prefix"
			criteriaValue = *selector.Criteria.PathPrefix
		case selector.Criteria.PathRegex != nil:
			criteriaName = "path_prefix"
			criteriaValue = *selector.Criteria.PathRegex
		case selector.Criteria.Postback != nil:
			criteriaName = "path_prefix"
			criteriaValue = *selector.Criteria.Postback
		default:
			fmt.Errorf("Error handling ABP site, selector is invalid, no valid criteria")
			continue
		}
		ret[i] = map[string]interface{}{
			"criteria_type":          criteriaName,
			"criteria_value":         criteriaValue,
			"analysis_rate_limiting": selector.AnalysisSettings.RateLimiting,
		}
		if selector.PolicyID != nil {
			ret[i]["policy_id"] = *selector.PolicyID
		}
		if selector.DerivedID != nil {
			ret[i]["derived_id"] = *selector.DerivedID
		}
	}

	return ret
}

func resourceABPSiteUpdate(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	var site ABPUpdateSite
	fillABPSiteData(d, &site)

	log.Printf("[DEBUG] ABP update site, created site for request: %v", site)

	client := m.(*Client)
	if _, err := client.updateABPSite(ID, &site); err != nil {
		return err
	}

	return resourceABPSiteRead(d, m)
}

func resourceABPSiteCreate(d *schema.ResourceData, m interface{}) error {
	var site ABPCreateSite
	fillABPSiteData(d, &site)
	site.Selectors = []ABPSiteSelector{}

	log.Printf("[DEBUG] ABP create site, created site for request: %v", site)

	client := m.(*Client)
	if s, err := client.createABPSite(&site); err != nil {
		return err
	} else {
		d.SetId(s.ID)
	}

	return resourceABPSiteRead(d, m)
}

func resourceABPSiteDelete(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	client := m.(*Client)

	log.Printf("[DEBUG] ABP delete site %s", ID)

	return client.deleteABPSite(ID)
}

func fillABPSiteData(d *schema.ResourceData, out interface{}) error {
	name := d.Get("name").(string)
	selectors := d.Get("selectors").([]interface{})
	site := ABPUpdateSite{}
	site.Name = name
	for _, s := range selectors {
		selector := s.(map[string]interface{})
		policyID := selector["policy_id"]
		criteriaName := selector["criteria_type"]
		criteriaValue := selector["criteria_value"]
		rateLimiting := selector["analysis_rate_limiting"]

		var siteSelector ABPSiteSelector
		p := policyID.(string)
		if len(p) > 0 {
			siteSelector.PolicyID = &p
		}
		siteSelector.AnalysisSettings.RateLimiting = rateLimiting.(string)
		cv := criteriaValue.(string)
		cn := criteriaName.(string)
		switch cn {
		case ABPSelectorCriteriaPostback:
			siteSelector.Criteria.Postback = &cv
		case ABPSelectorCriteriaPathPrefix:
			siteSelector.Criteria.PathPrefix = &cv
		case ABPSelectorCriteriaPathRegex:
			siteSelector.Criteria.PathRegex = &cv
		}

		site.Selectors = append(site.Selectors, siteSelector)
	}

	switch out.(type) {
	case *ABPCreateSite:
		mxName, mxNameSet := d.GetOk("mx_hostname_id")
		log.Printf("[DEBUG] ABP create site ID %s: %s %v %v", d.Id(), name, mxName, selectors)
		st := out.(*ABPCreateSite)
		if mxNameSet {
			s := mxName.(string)
			st.MxHostnameID = &s
		}
		st.ABPUpdateSite = site
	case *ABPUpdateSite:
		log.Printf("[DEBUG] ABP update site ID %s: %s %v", d.Id(), name, selectors)
		st := out.(*ABPUpdateSite)
		*st = site
	default:
		return fmt.Errorf("error got unsupported type to update with ABP site data: %v", out)
	}
	return nil
}

func compareSelectorsToSchemaResource(selectors []ABPSiteSelector, schemaSelectors []map[string]interface{}) bool {
	log.Printf("ABP Comparing selectors lists: %v VS %v", selectors, schemaSelectors)
	if len(selectors) != len(schemaSelectors) {
		return false
	}
	return reflect.DeepEqual(getSelectorsList(selectors), schemaSelectors)
}

func compareSiteToSchemaResource(d *schema.ResourceData, site *ABPSiteResponse) bool {
	name := d.Get("name").(string)
	var mxHostname, siteMXHostname string
	mxHostnameRes, ok := d.GetOk("mx_hostname_id")
	if ok {
		mxHostname = mxHostnameRes.(string)
	} else {
		mxHostname = ""
	}
	if site.MxHostnameID != nil {
		siteMXHostname = *site.MxHostnameID
	} else {
		siteMXHostname = ""
	}
	//selectors := d.Get("selectors").([]map[string]interface{})
	selectorsResList := d.Get("selectors").([]interface{})
	selectors := make([]map[string]interface{}, len(selectorsResList))
	for i, selector := range selectorsResList {
		selectors[i] = selector.(map[string]interface{})
	}

	log.Printf("ABP Compare site, comparing:\n %s %s %v\nTo:\n %s %s %v", name, mxHostname, selectors, site.Name, site.MxHostnameID, site.Selectors)

	if name != site.Name || mxHostname != siteMXHostname || len(selectors) != len(site.Selectors) {
		return false
	}

	return compareSelectorsToSchemaResource(site.Selectors, selectors)
}

func findSiteByName(client *Client, d *schema.ResourceData) *ABPSiteResponse {
	log.Printf("ABP finding site by name")
	accountID, err := client.GetABPAccountID()
	if err != nil {
		return nil
	}
	log.Printf("ABP finding site by name, account ID to use: %s", accountID)
	sites, err := client.GetABPSitesForAccount(accountID)
	if err != nil {
		fmt.Errorf("error while getting sites list: %s", err)
		return nil
	}
	log.Printf("ABP Got sites for account: %v", sites)
	for _, site := range sites {
		log.Printf("ABP site: %v", site)
		if compareSiteToSchemaResource(d, &site) {
			return &site
		}
	}

	return nil
}
