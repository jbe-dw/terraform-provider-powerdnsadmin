package powerdnsadmin

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	pdnsaclient "github.com/jbe-dw/go-powerdns-admin/client"
	"github.com/jbe-dw/go-powerdns-admin/client/apikey"
	"github.com/jbe-dw/go-powerdns-admin/models"
)

func resourcePDNSAdminAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourcePDNSAdminAPIKeyCreate,
		Read:   resourcePDNSAdminAPIKeyRead,
		Update: resourcePDNSAdminAPIKeyUpdate,
		Delete: resourcePDNSAdminAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"role": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.Title(strings.ToLower(val.(string)))
				},
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := strings.Title(strings.ToLower(val.(string)))
					if v != "Administrator" && v != "Operator" && v != "User" {
						errs = append(errs, fmt.Errorf("%q must be any of 'Administrator', 'Operator' or 'User', got: %q", key, v))
					}
					return
				},
			},

			"domains": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					role := strings.Title(strings.ToLower(d.Get("role").(string)))
					if role != "User" {
						return true
					}
					return false
				},
			},

			"plain_text_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hashed_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePDNSAdminAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Creating PowerDNS Admin API Key")

	role := d.Get("role").(string)
	setDomains := d.Get("domains").(*schema.Set).List()

	if role == "User" && len(setDomains) == 0 {
		return fmt.Errorf("API Key with a User role must have at least one domain")
	} else if role != "User" && len(setDomains) > 0 {
		setDomains = make([]interface{}, 0)
	}

	var domains models.PDNSAdminZones
	for _, domain := range setDomains {
		domains = append(domains, &models.PDNSAdminZonesItems{Name: domain.(string)})
	}

	Apikey := &models.APIKey{
		Description: d.Get("description").(string),
		Domains:     domains,
		Role:        &models.PDNSAdminRole{Name: role},
	}

	resource := apikey.NewAPIGenerateApikeyParams().WithApikey(Apikey)
	createdAPIKey, err := client.Apikey.APIGenerateApikey(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(createdAPIKey.Payload.ID, 10))
	d.Set("plain_text_key", createdAPIKey.Payload.PlainKey)
	return resourcePDNSAdminAPIKeyRead(d, meta)
}

func resourcePDNSAdminAPIKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Reading PowerDNS Admin API Key: %s", d.Id())

	ApikeyID, err := strconv.ParseInt(d.Id(), 10, 64)
	resource := apikey.NewAPIGetApikeyByIDParams().WithApikeyID(ApikeyID)
	APIKey, err := client.Apikey.APIGetApikeyByID(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't fetch PowerDNS Admin API Key (%s): %s", d.Id(), err)
	}

	d.Set("description", APIKey.Payload.Description)
	d.Set("role", APIKey.Payload.Role.Name)
	d.Set("hashed_key", APIKey.Payload.Key)

	var domains []string
	if d.Get("role").(string) == "User" && len(APIKey.Payload.Domains) > 0 {
		for _, domainItem := range APIKey.Payload.Domains {
			domains = append(domains, domainItem.Name)
		}
	}

	d.Set("domains", domains)

	return nil
}

func resourcePDNSAdminAPIKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	if d.HasChange("description") || d.HasChange("domains") || d.HasChange("role") {

		role := d.Get("role").(string)
		setDomains := d.Get("domains").(*schema.Set).List()

		_, n := d.GetChange("role")
		newRole := strings.Title(strings.ToLower(n.(string)))

		if d.HasChange("role") && newRole == "User" && len(setDomains) == 0 {
			return fmt.Errorf("API Key with a User role must have at least one domain")
		}

		if newRole != "User" {
			setDomains = make([]interface{}, 0)
			setDomains = append(setDomains, "\"\"")
		}

		var domains models.PDNSAdminZones
		for _, domain := range setDomains {
			domains = append(domains, &models.PDNSAdminZonesItems{Name: domain.(string)})
		}

		Apikey := &models.APIKey{
			Description: d.Get("description").(string),
			Domains:     domains,
			Role:        &models.PDNSAdminRole{Name: role},
		}

		ApikeyID, err := strconv.ParseInt(d.Id(), 10, 64)

		resource := apikey.NewAPIUpdateApikeyParams().WithApikeyID(ApikeyID).WithApikey(Apikey)
		updatedAPIKey, err := client.Apikey.APIUpdateApikey(resource, nil)
		if err != nil {
			return err
		}

		if updatedAPIKey == nil {
			return fmt.Errorf("An unknown error occured while updating API Key %q", d.Id())
		}
	}
	return resourcePDNSAdminAPIKeyRead(d, meta)
}

func resourcePDNSAdminAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Deleting PowerDNS Admin API Key: %s", d.Id())

	ApikeyID, err := strconv.ParseInt(d.Id(), 10, 64)
	resource := apikey.NewAPIDeleteApikeyParams().WithApikeyID(ApikeyID)
	APIKey, err := client.Apikey.APIDeleteApikey(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't delete PowerDNS Admin API Key (%s): %s", d.Id(), err)
	}

	if APIKey == nil {
		return fmt.Errorf("An unknown error occured while deleting API Key (%s): %s", d.Id(), err)
	}
	return nil
}
