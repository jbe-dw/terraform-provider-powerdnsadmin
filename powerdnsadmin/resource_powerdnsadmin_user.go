package powerdnsadmin

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	pdnsaclient "github.com/jbe-dw/go-powerdns-admin/client"
	"github.com/jbe-dw/go-powerdns-admin/client/account"
	"github.com/jbe-dw/go-powerdns-admin/client/user"
	"github.com/jbe-dw/go-powerdns-admin/models"
)

func resourcePDNSAdminUser() *schema.Resource {
	return &schema.Resource{
		Create: resourcePDNSAdminUserCreate,
		Read:   resourcePDNSAdminUserRead,
		Update: resourcePDNSAdminUserUpdate,
		Delete: resourcePDNSAdminUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
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

			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"firstname": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"lastname": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"role": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.Title(strings.ToLower(val.(string)))
				},
				Default: "User",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := strings.Title(strings.ToLower(val.(string)))
					if v != "Administrator" && v != "Operator" && v != "User" {
						errs = append(errs, fmt.Errorf("%q must be any of 'Administrator', 'Operator' or 'User', got: %q", key, v))
					}
					return
				},
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9]"+
						"(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9]"+
						"(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
					"is not a valid email"),
			},
			"accounts": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"external": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"userid": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourcePDNSAdminUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Creating PowerDNS Admin User")

	// Create the User Object
	UserName := d.Get("username").(string)
	role := strings.Title(strings.ToLower(d.Get("role").(string)))
	email := d.Get("email").(string)

	User := &models.APICreateUserParamsBody{
		Email:     &email,
		Firstname: d.Get("firstname").(string),
		Lastname:  d.Get("lastname").(string),
		RoleName:  role,
		Username:  &UserName,
	}

	if d.Get("external") == true {
		password := "*"
		User.Password = password
	} else {
		if d.Get("password") == nil {
			return fmt.Errorf("PowerDNS Admin User %s must have a password", d.Get("username"))
		}
		password := d.Get("password").(string)
		User.PlainTextPassword = password
	}

	resource := user.NewAPICreateUserParams().WithUser(User)
	createdUser, err := client.User.APICreateUser(resource, nil)
	if err != nil {
		return err
	}

	userID := createdUser.Payload.ID

	// Link the User to the accounts
	userAccounts := d.Get("accounts").(*schema.Set).List()
	if userAccounts != nil && len(userAccounts) > 0 {

		log.Printf("[DEBUG] Reading PowerDNS Admin Accounts")

		resource := account.NewAPIListAccountsParams()
		AccountList, err := client.Account.APIListAccounts(resource, nil)
		if err != nil {
			return fmt.Errorf("Couldn't fetch PowerDNS Admin Accounts: %s", err)
		} else if len(AccountList.Payload) == 0 {
			return fmt.Errorf("PowerDNS Admin has no account while you requested some")
		}

		/*AccountMap := make(map[string]int64)
		AccountNameList := make([]string, len(AccountList.Payload))
		for _, account := range AccountList.Payload {
			AccountMap[account.Name] = account.ID
			AccountNameList = append(AccountNameList, account.Name)
		}*/
		AccountMap, err := pDNSAdminAccountList(meta)
		if err != nil {
			return err
		}

		for _, userAccount := range userAccounts {
			accountName := userAccount.(string)
			accountID, ok := AccountMap[accountName]
			if ok == false {
				// return fmt.Errorf("The account %s was not found", accountName)
				continue
			}
			// Link the user to the account
			res := account.NewAPIAddAccountUserParams().WithAccountID(accountID).WithUserID(userID)
			_, err = client.Account.APIAddAccountUser(res, nil)
			if err != nil {
				return fmt.Errorf("Couldn't link PowerDNS Admin User %s to account %s: %s",
					UserName, accountName, err)
			}
		}

	}

	d.SetId(createdUser.Payload.Username)
	d.Set("userid", userID)
	return resourcePDNSAdminUserRead(d, meta)
}

func resourcePDNSAdminUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	// Read User
	UserName := d.Id()
	log.Printf("[DEBUG] Reading PowerDNS Admin User: %s", UserName)
	resource := user.NewAPIGetUserParams().WithUsername(UserName)
	User, err := client.User.APIGetUser(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't fetch PowerDNS Admin User (%s): %s", UserName, err)
	}

	d.Set("username", User.Payload.Username)
	d.Set("firstname", User.Payload.Firstname)
	d.Set("lastname", User.Payload.Lastname)
	d.Set("role", User.Payload.Role.Name)
	d.Set("userid", User.Payload.ID)

	// Read Accounts
	var accounts []string
	for _, account := range User.Payload.Accounts {
		accounts = append(accounts, account.Name)
	}

	d.Set("accounts", accounts)

	return nil
}

func resourcePDNSAdminUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	log.Printf("[DEBUG] Updating PowerDNS Admin User")

	// Update User
	if d.HasChange("firstname") || d.HasChange("lastname") || d.HasChange("email") || d.HasChange("role") {

		log.Printf("[DEBUG] Updating PowerDNS Admin User parameters")

		UserID := int64(d.Get("userid").(int))
		role := strings.Title(strings.ToLower(d.Get("role").(string)))

		User := &models.APIUpdateUserParamsBody{
			Email:     d.Get("email").(string),
			Firstname: d.Get("firstname").(string),
			Lastname:  d.Get("lastname").(string),
			RoleName:  role,
		}

		resource := user.NewAPIUpdateUserParams().WithUserID(UserID).WithUser(User)
		updatedUser, err := client.User.APIUpdateUser(resource, nil)
		if err != nil {
			return err
		}

		if updatedUser == nil {
			return fmt.Errorf("An unknown error occured while updating User %s", UserID)
		}
	}

	// Update Account
	if d.HasChange("accounts") {

		log.Printf("[DEBUG] Updating PowerDNS Admin User account links")

		// Refresh
		resource := user.NewAPIGetUserParams().WithUsername(d.Id())
		User, err := client.User.APIGetUser(resource, nil)
		if err != nil {
			return fmt.Errorf("Couldn't fetch PowerDNS Admin User (%s) latest values: %s", d.Id(), err)
		}

		userID := int64(d.Get("userid").(int))
		userAccounts := d.Get("accounts").(*schema.Set).List()
		obsoleteAccounts := make(map[string]int64)
		var missingAccounts []string
		match := false

		// Filter out linked accounts that are not needed
		for _, account := range User.Payload.Accounts {
			match = false
			for _, reqAccount := range userAccounts {
				if account.Name == reqAccount.(string) {
					match = true
					continue
				}
			}
			if match == false {
				obsoleteAccounts[account.Name] = account.ID
			}
		}

		// Filter out accounts that are needed but not linked yet
		for _, reqAccount := range userAccounts {
			match = false
			for _, account := range User.Payload.Accounts {
				if reqAccount.(string) == account.Name {
					match = true
					continue
				}
			}
			if match == false {
				missingAccounts = append(missingAccounts, reqAccount.(string))
			}
		}

		if len(obsoleteAccounts) > 0 {
			for obsAccountName, obsAccountID := range obsoleteAccounts {
				res := account.NewAPIRemoveAccountUserParams().WithAccountID(obsAccountID).WithUserID(userID)
				_, err = client.Account.APIRemoveAccountUser(res, nil)
				if err != nil {
					return fmt.Errorf("Couldn't unlink PowerDNS Admin User %s (%s) to account %s (%s): %s",
						d.Id(), userID, obsAccountName, obsAccountID, err)
				}
			}
		}

		if len(missingAccounts) > 0 {
			AccountMap, err := pDNSAdminAccountList(meta)
			if err != nil {
				return err
			}

			for _, misAccount := range missingAccounts {
				accountID, ok := AccountMap[misAccount]
				if ok == false {
					return fmt.Errorf("The account %s was not found", misAccount)
				}

				res := account.NewAPIAddAccountUserParams().WithAccountID(accountID).WithUserID(userID)
				_, err = client.Account.APIAddAccountUser(res, nil)

				if err != nil {
					return fmt.Errorf("Couldn't link PowerDNS Admin User %s to account %s: %s",
						d.Id(), misAccount, err)
				}
			}

		}
	}
	return resourcePDNSAdminUserRead(d, meta)
}

func resourcePDNSAdminUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pdnsaclient.Pdnsadmin)

	UserID := int64(d.Get("userid").(int))

	log.Printf("[DEBUG] Deleting PowerDNS Admin User: %s", UserID)

	resource := user.NewAPIDeleteUserParams().WithUserID(UserID)
	User, err := client.User.APIDeleteUser(resource, nil)
	if err != nil {
		return fmt.Errorf("Couldn't delete PowerDNS Admin User (%s): %s", UserID, err)
	}

	if User == nil {
		return fmt.Errorf("An unknown error occured while deleting User (%s): %s", UserID, err)
	}
	return nil
}

func pDNSAdminAccountList(meta interface{}) (AccountMap map[string]int64, err error) {
	client := meta.(*pdnsaclient.Pdnsadmin)

	resource := account.NewAPIListAccountsParams()
	AccountList, err := client.Account.APIListAccounts(resource, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't fetch PowerDNS Admin Accounts: %s", err)
	} else if len(AccountList.Payload) == 0 {
		return nil, fmt.Errorf("PowerDNS Admin has no account while you requested some")
	}

	AccountMap = make(map[string]int64)
	for _, account := range AccountList.Payload {
		AccountMap[account.Name] = account.ID
	}
	return AccountMap, nil
}
