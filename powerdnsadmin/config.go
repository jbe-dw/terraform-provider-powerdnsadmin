package powerdnsadmin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	rc "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/pathorcontents"
	pdnsaclient "github.com/jbe-dw/go-powerdns-admin/client"
)

// Config describes de configuration interface of this provider
type Config struct {
	Host          string
	User          string
	Password      string
	Scheme        string
	InsecureHTTPS bool
	CACertificate string
}

// Client returns a new client for accessing PowerDNS
func (c *Config) Client() (*pdnsaclient.Pdnsadmin, error) {

	// TLS
	if c.Scheme == "https" {
		tlsConfig := &tls.Config{}

		if c.CACertificate != "" {

			caCert, _, err := pathorcontents.Read(c.CACertificate)
			if err != nil {
				return nil, fmt.Errorf("Error reading CA Cert: %s", err)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(caCert))
			tlsConfig.RootCAs = caCertPool
		}

		tlsConfig.InsecureSkipVerify = c.InsecureHTTPS
	}

	defaultScheme := []string{c.Scheme}
	t := rc.New(c.Host, pdnsaclient.DefaultBasePath, defaultScheme)
	t.DefaultAuthentication = rc.BasicAuth(c.User, c.Password)
	client := pdnsaclient.New(t, strfmt.Default)

	log.Printf("[INFO] PowerDNS Admin Client configured for server %s", c.Host)

	return client, nil
}
