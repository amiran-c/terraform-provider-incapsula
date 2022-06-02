package incapsula

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

func resourceABPAccount() *schema.Resource {
	return &schema.Resource{
		Create: nil,
		Read:   resourceABPAccountRead,
		Update: nil,
		Delete: resourceCSPSiteDomainDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				keyParts := strings.Split(d.Id(), "-")
				if len(keyParts) != 5 {
					return nil, fmt.Errorf("Error parsing ID, actual value: %s, expected ABP account ID\n", d.Id())
				}
				log.Printf("[DEBUG] Import ABP account ID %s", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			// Required Arguments
			"my_account_id": {
				Description: "Numeric identifier of the account in MY.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"name": {
				Description: "The name of the account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceABPAccountRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	ID := d.Id()
	log.Printf("[DEBUG] Reading ABP account ID: %s", ID)

	abpAccountResponse, err := client.GetABPAccount(ID)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("[ERROR] Could not read ABP account %s: %s", ID, err)
	}
	d.Set("my_account_id", abpAccountResponse.MYAccountID)
	d.Set("name", abpAccountResponse.Name)
	d.SetId(abpAccountResponse.ID)

	return nil
}

func resourceABPAccountUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Can't UPDATE ABP Account. ID %s.", d.Id())

	return resourceABPAccountRead(d, m)
}

func resourceABPAccountDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Can't DELETE ABP Account. ID %s.", d.Id())

	return nil
}
