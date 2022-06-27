package incapsula

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

func resourceABPPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceABPPolicyCreate,
		Read:   resourceABPPolicyRead,
		Update: resourceABPPolicyUpdate,
		Delete: resourceABPPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyParts := strings.Split(d.Id(), "-")
				if len(keyParts) == 5 {
					return importABPPolicyResourcesByID(d, meta.(*Client))
				} else {
					return importABPPolicyResourcesByName(d, meta.(*Client))
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
			"name": {
				Description: "The name of the policy.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "An optional user-defined description for this policy.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"directives": {
				Description: "List of directives, a directive is a set of conditions coupled with an action to perform when traffic matches said conditions.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func importABPPolicyResourcesByID(d *schema.ResourceData, client *Client) ([]*schema.ResourceData, error) {
	if _, err := client.GetABPPolicy(d.Id()); err == nil {
		return []*schema.ResourceData{d}, nil
	}
	return nil, fmt.Errorf("ABP site ID %s not found, can't import", d.Id())
}

func importABPPolicyResourcesByName(d *schema.ResourceData, client *Client) ([]*schema.ResourceData, error) {
	// Save the policy name from the ID to use later for finding the policies
	policyName := d.Id()
	found := false

	policies, err := client.GetABPPoliciesForAccountWithoutID()
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] ABP policy import, fetched %d ABP policies: %v", len(policies), policies)

	// Create a list of resources and add by matching policy name
	results := []*schema.ResourceData{d}
	for _, policy := range policies {
		if policyName == policy.Name {
			log.Printf("[DEBUG] ABP policy import adding new policy ID (%s) , name (%s)", policy.ID, policy.Name)
			var cd *schema.ResourceData
			if found == false {
				d.SetId(policy.ID)
				found = true
				log.Printf("[DEBUG] ABP policy import found wanted resource %v , results: %v", cd, results)
			} else {
				cd = resourceABPSite().Data(nil)
				cd.SetId(policy.ID)
				cd.SetType("incapsula_abp_policy")
				results = append(results, cd)
				log.Printf("[DEBUG] ABP policy import created new resource %v , results: %v", cd, results)
			}
		} else {
			log.Printf("[DEBUG] ABP policy import skipping current policy name (%s) because it's not (%s)", policy.Name, policyName)
		}
	}
	log.Printf("[DEBUG] ABP policy import, importing %d resources: %v", len(results), results)

	return results, nil
}

func resourceABPPolicyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	ID := d.Id()

	keyParts := strings.Split(ID, "-")
	if len(keyParts) != 5 {
		d.SetId("")
		return fmt.Errorf("ABP read policy, ID was not resolved on import, can't fetch site (%s)", ID)
	}
	log.Printf("[DEBUG] ABP read policy ID: %s", ID)

	abpPolicyResponse, err := client.GetABPPolicy(ID)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("error reading ABP policy ID %s: %s", ID, err)
	} else {
		log.Printf("[DEBUG] ABP read policy, got policy: %v", abpPolicyResponse)
	}
	d.SetId(abpPolicyResponse.ID)
	d.Set("name", abpPolicyResponse.Name)
	d.Set("account_id", abpPolicyResponse.AccountID)
	if abpPolicyResponse.Description != nil {
		d.Set("description", abpPolicyResponse.Description)
	}

	if len(abpPolicyResponse.Directives) > 0 {
		if err := d.Set("directives", getDirectivesList(abpPolicyResponse.Directives)); err != nil {
			return fmt.Errorf("error handling ABP policy, error setting directives: %s", err)
		}
	}

	return nil
}

func getDirectivesList(directives []ABPPolicyDirective) []map[string]interface{} {
	log.Printf("directives: %v", directives)
	if len(directives) == 0 {
		return nil
	}
	ret := make([]map[string]interface{}, len(directives))
	log.Printf("directives: %v", ret)

	for i, directive := range directives {
		log.Printf("directive index %d: %v", i, directive)
		ret[i] = map[string]interface{}{
			"action":       directive.Action,
			"condition_id": *directive.ConditionID,
		}
	}

	return ret
}

func resourceABPPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	var policy ABPCreateUpdatePolicy
	fillABPPolicyData(d, &policy)

	log.Printf("[DEBUG] ABP update policy, created policy for request: %v", policy)

	client := m.(*Client)
	if _, err := client.updateABPPolicy(ID, &policy); err != nil {
		return err
	}

	return resourceABPPolicyRead(d, m)
}

func resourceABPPolicyCreate(d *schema.ResourceData, m interface{}) error {
	var policy ABPCreateUpdatePolicy
	fillABPPolicyData(d, &policy)

	log.Printf("[DEBUG] ABP create policy, created policy for request: %v", policy)

	client := m.(*Client)
	if s, err := client.createABPPolicy(&policy); err != nil {
		return err
	} else {
		d.SetId(s.ID)
	}

	return resourceABPPolicyRead(d, m)
}

func resourceABPPolicyDelete(d *schema.ResourceData, m interface{}) error {
	ID := d.Id()
	client := m.(*Client)

	log.Printf("[DEBUG] ABP delete site %s", ID)

	return client.deleteABPPolicy(ID)
}

func fillABPPolicyData(d *schema.ResourceData, out *ABPCreateUpdatePolicy) error {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	directives := d.Get("directives").([]interface{})
	out.Name = name
	if len(description) > 0 {
		out.Description = &description
	}
	for _, dir := range directives {
		directive := dir.(map[string]interface{})
		var policyDirective ABPPolicyDirective

		conditionID := directive["condition_id"].(string)
		policyDirective.ConditionID = &conditionID
		policyDirective.Action = directive["action"].(string)
		out.Directives = append(out.Directives, policyDirective)
	}

	return nil
}
