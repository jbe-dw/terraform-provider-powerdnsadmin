package powerdnsadmin

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	pdnsaclient "github.com/jbe-dw/go-powerdns-admin/client"
	"github.com/jbe-dw/go-powerdns-admin/client/account"
	"github.com/jbe-dw/go-powerdns-admin/client/user"
	"github.com/jbe-dw/go-powerdns-admin/models"
)

func resourcePDNSAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourcePDNSAdminAccountCreate,
		Read:   resourcePDNSAdminAccountRead,
		Update: resourcePDNSAdminAccountUpdate,
		Delete: resourcePDNSAdminAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.ToLower(old) == strings.ToLower(new) {
						return true
					}
					return false
				},
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[a-z0-9]`), "Must be alphanumeric lowercase"),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"contact_email": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9]"+
						"(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9]"+
						"(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
					"is not a valid email"),
			},

			"contact_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourcePDNSAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Creating PowerDNS Admin Account")

	AccountName := d.Get("name").(string)
	Account := &models.APICreateAccountParamsBody{
		Contact:     d.Get("contact_name").(string),
		Description: d.Get("description").(string),
		Mail:        d.Get("contact_email").(string),
		Name:        &AccountName,
	}

	resource := account.NewAPICreateAccountParams().WithAccount(Account)
	createdAccount, err := client.Account.APICreateAccount(resource, nil)
	if err != nil {
		return err
	}

	//d.SetId(strconv.FormatInt(createdAccount.Payload.ID, 10))
	d.SetId(createdAccount.Payload.Name)
	resourcePDNSAdminAccountRead(d, meta)

	return nil
}

func resourcePDNSAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	AccountName := d.Id()
	log.Printf("[DEBUG] Reading PowerDNS Admin Account: %s", AccountName)
	resource := user.NewAPIGetAccountByNameParams().WithAccountName(AccountName)
	Account, err := client.User.APIGetAccountByName(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't fetch PowerDNS Admin Account (%s): %s", AccountName, err)
	}

	d.Set("description", Account.Payload.Description)
	d.Set("contact_name", Account.Payload.Contact)
	d.Set("contact_mail", Account.Payload.Mail)
	d.Set("name", Account.Payload.Name)

	return nil
}

func resourcePDNSAdminAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	if d.HasChange("description") || d.HasChange("contact_name") || d.HasChange("contact_email") {

		AccountID, err := strconv.ParseInt(d.Id(), 10, 64)

		Account := &models.APIUpdateAccountParamsBody{
			Contact:     d.Get("contact_name").(string),
			Description: d.Get("description").(string),
			Mail:        d.Get("contact_email").(string),
		}

		resource := user.NewAPIUpdateAccountParams().WithAccountID(AccountID).WithAccount(Account)
		updatedAccount, err := client.User.APIUpdateAccount(resource, nil)
		if err != nil {
			return err
		}

		if updatedAccount == nil {
			return fmt.Errorf("An unknown error occured while updating Acount %s", AccountID)
		}
		resourcePDNSAdminAccountRead(d, meta)
	}
	return nil
}

func resourcePDNSAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	AccountID, err := strconv.ParseInt(d.Id(), 10, 64)

	log.Printf("[DEBUG] Deleting PowerDNS Admin Account: %s", AccountID)

	resource := user.NewAPIDeleteAccountParams().WithAccountID(AccountID)
	Account, err := client.User.APIDeleteAccount(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't delete PowerDNS Admin Account (%s): %s", AccountID, err)
	}

	if Account == nil {
		return fmt.Errorf("An unknown error occured while deleting Account (%s): %s", AccountID, err)
	}
	return nil
}
