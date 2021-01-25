package powerdnsadmin

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a schema.Provider for PowerDNSAdmin.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_USER", nil),
				Description: "Basic Auth Username",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_PASSWORD", nil),
				Description: "Basic Auth Password",
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_HOST", nil),
				Description: "PowerDNS Admin server Address",
			},
			"scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_SCHEME", "http"),
				Description: "Sheme used to connect PowerDNS Admin.",
			},
			"insecure_https": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_INSECURE_HTTPS", false),
				Description: "Disable verification of the PowerDNS server's TLS certificate",
			},
			"ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PDNSA_CACERT", ""),
				Description: "Content or path of a Root CA to be used to verify PowerDNS's SSL certificate",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"powerdnsadmin_apikey":  resourcePDNSAdminAPIKey(),
			"powerdnsadmin_account": resourcePDNSAdminAccount(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	config := Config{
		User:          data.Get("user").(string),
		Password:      data.Get("password").(string),
		Host:          data.Get("host").(string),
		Scheme:        data.Get("scheme").(string),
		InsecureHTTPS: data.Get("insecure_https").(bool),
		CACertificate: data.Get("ca_certificate").(string),
	}

	return config.Client()
}
