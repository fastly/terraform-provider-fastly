package backend

import (
	"context"
	"sort"

	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultPort                = 80
	DefaultAutoLoadbalance     = false
	DefaultBetweenBytesTimeout = 10000
	DefaultConnectTimeout      = 1000
	DefaultErrorThreshold      = 0
	DefaultFirstByteTimeout    = 15000
	DefaultHealthCheck         = ""
	DefaultMaxConn             = 200
	DefaultMaxLifetime         = 0
	DefaultMaxTLSVersion       = ""
	DefaultMaxUse              = 0
	DefaultMinTLSVersion       = ""
	DefaultOverrideHost        = ""
	DefaultPreferIPv6          = false
	DefaultRequestCondition    = ""
	DefaultShareKey            = ""
	DefaultShield              = ""
	DefaultSSLCACert           = ""
	DefaultSSLCertHostname     = ""
	DefaultSSLCheckCert        = true
	DefaultSSLCiphers          = ""
	DefaultSSLClientCert       = ""
	DefaultSSLClientKey        = ""
	DefaultSSLSNIHostname      = ""
	DefaultUseSSL              = false
	DefaultWeight              = 100
)

type NestedModel struct {
	Name                types.String `tfsdk:"name"`
	Address             types.String `tfsdk:"address"`
	Port                types.Int64  `tfsdk:"port"`
	Comment             types.String `tfsdk:"comment"`
	AutoLoadbalance     types.Bool   `tfsdk:"auto_loadbalance"`
	BetweenBytesTimeout types.Int64  `tfsdk:"between_bytes_timeout"`
	ConnectTimeout      types.Int64  `tfsdk:"connect_timeout"`
	ErrorThreshold      types.Int64  `tfsdk:"error_threshold"`
	FirstByteTimeout    types.Int64  `tfsdk:"first_byte_timeout"`
	HealthCheck         types.String `tfsdk:"healthcheck"`
	KeepaliveTime       types.Int64  `tfsdk:"keepalive_time"`
	MaxConn             types.Int64  `tfsdk:"max_conn"`
	MaxLifetime         types.Int64  `tfsdk:"max_lifetime"`
	MaxTLSVersion       types.String `tfsdk:"max_tls_version"`
	MaxUse              types.Int64  `tfsdk:"max_use"`
	MinTLSVersion       types.String `tfsdk:"min_tls_version"`
	OverrideHost        types.String `tfsdk:"override_host"`
	PreferIPv6          types.Bool   `tfsdk:"prefer_ipv6"`
	RequestCondition    types.String `tfsdk:"request_condition"`
	ShareKey            types.String `tfsdk:"share_key"`
	Shield              types.String `tfsdk:"shield"`
	SSLCACert           types.String `tfsdk:"ssl_ca_cert"`
	SSLCertHostname     types.String `tfsdk:"ssl_cert_hostname"`
	SSLCheckCert        types.Bool   `tfsdk:"ssl_check_cert"`
	SSLCiphers          types.String `tfsdk:"ssl_ciphers"`
	SSLClientCert       types.String `tfsdk:"ssl_client_cert"`
	SSLClientKey        types.String `tfsdk:"ssl_client_key"`
	SSLSNIHostname      types.String `tfsdk:"ssl_sni_hostname"`
	UseSSL              types.Bool   `tfsdk:"use_ssl"`
	Weight              types.Int64  `tfsdk:"weight"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Address) == service.StringValue(other.Address) &&
		service.Int64Value(n.Port) == service.Int64Value(other.Port) &&
		service.StringValue(n.Comment) == service.StringValue(other.Comment) &&
		service.BoolValue(n.AutoLoadbalance) == service.BoolValue(other.AutoLoadbalance) &&
		service.Int64Value(n.BetweenBytesTimeout) == service.Int64Value(other.BetweenBytesTimeout) &&
		service.Int64Value(n.ConnectTimeout) == service.Int64Value(other.ConnectTimeout) &&
		service.Int64Value(n.ErrorThreshold) == service.Int64Value(other.ErrorThreshold) &&
		service.Int64Value(n.FirstByteTimeout) == service.Int64Value(other.FirstByteTimeout) &&
		service.StringValue(n.HealthCheck) == service.StringValue(other.HealthCheck) &&
		service.Int64Value(n.KeepaliveTime) == service.Int64Value(other.KeepaliveTime) &&
		service.Int64Value(n.MaxConn) == service.Int64Value(other.MaxConn) &&
		service.Int64Value(n.MaxLifetime) == service.Int64Value(other.MaxLifetime) &&
		service.StringValue(n.MaxTLSVersion) == service.StringValue(other.MaxTLSVersion) &&
		service.Int64Value(n.MaxUse) == service.Int64Value(other.MaxUse) &&
		service.StringValue(n.MinTLSVersion) == service.StringValue(other.MinTLSVersion) &&
		service.StringValue(n.OverrideHost) == service.StringValue(other.OverrideHost) &&
		service.BoolValue(n.PreferIPv6) == service.BoolValue(other.PreferIPv6) &&
		service.StringValue(n.RequestCondition) == service.StringValue(other.RequestCondition) &&
		service.StringValue(n.ShareKey) == service.StringValue(other.ShareKey) &&
		service.StringValue(n.Shield) == service.StringValue(other.Shield) &&
		service.StringValue(n.SSLCACert) == service.StringValue(other.SSLCACert) &&
		service.StringValue(n.SSLCertHostname) == service.StringValue(other.SSLCertHostname) &&
		service.BoolValue(n.SSLCheckCert) == service.BoolValue(other.SSLCheckCert) &&
		service.StringValue(n.SSLCiphers) == service.StringValue(other.SSLCiphers) &&
		service.StringValue(n.SSLClientCert) == service.StringValue(other.SSLClientCert) &&
		service.StringValue(n.SSLClientKey) == service.StringValue(other.SSLClientKey) &&
		service.StringValue(n.SSLSNIHostname) == service.StringValue(other.SSLSNIHostname) &&
		service.BoolValue(n.UseSSL) == service.BoolValue(other.UseSSL) &&
		service.Int64Value(n.Weight) == service.Int64Value(other.Weight)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Name for this backend. Must be unique within the service.",
		},
		"address": schema.StringAttribute{
			Required:    true,
			Description: "An IPv4 address, IPv6 address, or hostname for the backend.",
		},
		"port": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultPort),
			Description: "The port number on which the backend responds. Default `80`.",
		},
		"comment": schema.StringAttribute{
			Optional:    true,
			Description: "Optional comment for the backend.",
		},
		"auto_loadbalance": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultAutoLoadbalance),
			Description: "Whether this backend should be included in automatic load balancing. CDN services only. Default `false`.",
		},
		"between_bytes_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultBetweenBytesTimeout),
			Description: "How long to wait between bytes in milliseconds. Default `10000`.",
		},
		"connect_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultConnectTimeout),
			Description: "How long to wait for a timeout in milliseconds. Default `1000`.",
		},
		"error_threshold": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultErrorThreshold),
			Description: "Number of errors to allow before the backend is marked as down. Default `0`.",
		},
		"first_byte_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultFirstByteTimeout),
			Description: "How long to wait for the first byte in milliseconds. Default `15000`.",
		},
		"healthcheck": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultHealthCheck),
			Description: "Name of a defined healthcheck to assign to this backend.",
		},
		"keepalive_time": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "How long in seconds to keep a persistent connection to the backend between requests.",
		},
		"max_conn": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultMaxConn),
			Description: "Maximum number of connections for this backend. Default `200`.",
		},
		"max_lifetime": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultMaxLifetime),
			Description: "Maximum time from creation, in milliseconds, that a pooled HTTP keepalive connection is eligible for reuse. `0` is treated as unlimited.",
		},
		"max_tls_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultMaxTLSVersion),
			Description: "Maximum allowed TLS version on SSL connections to this backend.",
		},
		"max_use": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultMaxUse),
			Description: "Maximum number of requests allowed over a single pooled HTTP keepalive connection. `0` is treated as unlimited.",
		},
		"min_tls_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultMinTLSVersion),
			Description: "Minimum allowed TLS version on SSL connections to this backend.",
		},
		"override_host": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultOverrideHost),
			Description: "Hostname to override the Host header.",
		},
		"prefer_ipv6": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultPreferIPv6),
			Description: "Prefer IPv6 connections to origins for hostname backends. Default `false` for CDN services.",
		},
		"request_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultRequestCondition),
			Description: "Name of a request condition which, if met, selects this backend.",
		},
		"share_key": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultShareKey),
			Description: "Value that, when shared across backends, enables those backends to share the same health check.",
		},
		"shield": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultShield),
			Description: "POP of the shield designated to reduce inbound load.",
		},
		"ssl_ca_cert": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSSLCACert),
			Description: "CA certificate attached to origin.",
		},
		"ssl_cert_hostname": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSSLCertHostname),
			Description: "Hostname used for certificate validation. Does not affect SNI.",
		},
		"ssl_check_cert": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultSSLCheckCert),
			Description: "Whether to strictly check SSL certificates. Default `true`.",
		},
		"ssl_ciphers": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSSLCiphers),
			Description: "Cipher list for TLS connections to this backend.",
		},
		"ssl_client_cert": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Default:     stringdefault.StaticString(DefaultSSLClientCert),
			Description: "Client certificate used when connecting to the backend.",
		},
		"ssl_client_key": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Default:     stringdefault.StaticString(DefaultSSLClientKey),
			Description: "Client key used when connecting to the backend.",
		},
		"ssl_sni_hostname": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSSLSNIHostname),
			Description: "Hostname used for SNI in the TLS handshake.",
		},
		"use_ssl": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultUseSSL),
			Description: "Whether to use SSL to reach the backend. Default `false`.",
		},
		"weight": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultWeight),
			Description: "Portion of traffic to send to this backend. Default `100`.",
		},
	}
}

func ResourceAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier.",
		},
		"service_id": schema.StringAttribute{
			Required:    true,
			Description: "Fastly service ID.",
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
		},
	}
	for k, v := range CommonAttributes() {
		attrs[k] = v
	}
	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Backends attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

func Normalize(input []NestedModel) []NestedModel {
	out := make([]NestedModel, len(input))
	copy(out, input)

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name.ValueString() < out[j].Name.ValueString()
	})

	return out
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	remote, err := client.ListBackends(ctx, &fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	result := make([]NestedModel, 0, len(remote))
	for _, b := range remote {
		result = append(result, FlattenToNestedModel(b))
	}

	return Normalize(result), nil
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	remote, err := client.ListBackends(ctx, &fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}

	desired = Normalize(desired)

	remoteByName := make(map[string]*fastly.Backend, len(remote))
	for _, b := range remote {
		remoteByName[fastly.ToValue(b.Name)] = b
	}

	desiredByName := make(map[string]NestedModel, len(desired))
	for _, b := range desired {
		desiredByName[b.Name.ValueString()] = b
	}

	// Delete backends no longer present.
	for name := range remoteByName {
		if _, ok := desiredByName[name]; !ok {
			err := client.DeleteBackend(ctx, &fastly.DeleteBackendInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
				Name:           name,
			})
			if errors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	// Create or update desired backends.
	for _, desiredBackend := range desired {
		name := desiredBackend.Name.ValueString()
		remoteBackend, exists := remoteByName[name]

		if !exists {
			input := BuildCreateInput(serviceID, version, desiredBackend)
			if _, err := client.CreateBackend(ctx, input); err != nil {
				return err
			}
			continue
		}

		remoteModel := FlattenToNestedModel(remoteBackend)
		if !desiredBackend.ModelsEqual(remoteModel) {
			input := BuildUpdateInput(serviceID, version, desiredBackend, &remoteModel, false)
			if _, err := client.UpdateBackend(ctx, input); err != nil {
				return err
			}
		}
	}

	return nil
}

func Equal(a, b []NestedModel) bool {
	a = Normalize(a)
	b = Normalize(b)

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].ModelsEqual(b[i]) {
			return false
		}
	}

	return true
}
