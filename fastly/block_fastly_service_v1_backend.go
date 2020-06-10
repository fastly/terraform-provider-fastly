package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var backendSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for this Backend",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "An IPv4, hostname, or IPv6 address for the Backend",
			},
			// Optional fields, defaults where they exist
			"auto_loadbalance": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Should this Backend be load balanced",
			},
			"between_bytes_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10000,
				Description: "How long to wait between bytes in milliseconds",
			},
			"connect_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1000,
				Description: "How long to wait for a timeout in milliseconds",
			},
			"error_threshold": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of errors to allow before the Backend is marked as down",
			},
			"first_byte_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15000,
				Description: "How long to wait for the first bytes in milliseconds",
			},
			"healthcheck": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The healthcheck name that should be used for this Backend",
			},
			"max_conn": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     200,
				Description: "Maximum number of connections for this Backend",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     80,
				Description: "The port number Backend responds on. Default 80",
			},
			"override_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hostname to override the Host header",
			},
			"request_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition, which if met, will select this backend during a request.",
			},
			"shield": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The POP of the shield designated to reduce inbound load.",
			},
			"use_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not to use SSL to reach the Backend",
			},
			"max_tls_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Maximum allowed TLS version on SSL connections to this backend.",
			},
			"min_tls_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Minimum allowed TLS version on SSL connections to this backend.",
			},
			"ssl_ciphers": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Comma sepparated list of ciphers",
			},
			"ssl_check_cert": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Be strict on checking SSL certs",
			},
			"ssl_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSL certificate hostname",
				Deprecated:  "Use ssl_cert_hostname and ssl_sni_hostname instead.",
			},
			"ssl_ca_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "CA certificate attached to origin.",
			},
			"ssl_cert_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSL certificate hostname for cert verification",
			},
			"ssl_sni_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSL certificate hostname for SNI verification",
			},
			"ssl_client_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSL certificate file for client connections to the backend.",
				Sensitive:   true,
			},
			"ssl_client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "SSL key file for client connections to backend.",
				Sensitive:   true,
			},

			"weight": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "The portion of traffic to send to a specific origins. Each origin receives weight/total of the traffic.",
			},
		},
	},
}
