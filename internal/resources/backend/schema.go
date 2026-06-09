package backend

import (
	"context"
	"sort"

	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			Default:     int64default.StaticInt64(80),
			Description: "The port number on which the backend responds. Default `80`.",
		},
		"comment": schema.StringAttribute{
			Optional:    true,
			Description: "Optional comment for the backend.",
		},
		"auto_loadbalance": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Whether this backend should be included in automatic load balancing. CDN services only. Default `false`.",
		},
		"between_bytes_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(10000),
			Description: "How long to wait between bytes in milliseconds. Default `10000`.",
		},
		"connect_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(1000),
			Description: "How long to wait for a timeout in milliseconds. Default `1000`.",
		},
		"error_threshold": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "Number of errors to allow before the backend is marked as down. Default `0`.",
		},
		"first_byte_timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(15000),
			Description: "How long to wait for the first byte in milliseconds. Default `15000`.",
		},
		"healthcheck": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Name of a defined healthcheck to assign to this backend.",
		},
		"keepalive_time": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "How long in seconds to keep a persistent connection to the backend between requests.",
		},
		"max_conn": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(200),
			Description: "Maximum number of connections for this backend. Default `200`.",
		},
		"max_lifetime": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "Maximum time from creation, in milliseconds, that a pooled HTTP keepalive connection is eligible for reuse. `0` is treated as unlimited.",
		},
		"max_tls_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Maximum allowed TLS version on SSL connections to this backend.",
		},
		"max_use": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "Maximum number of requests allowed over a single pooled HTTP keepalive connection. `0` is treated as unlimited.",
		},
		"min_tls_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Minimum allowed TLS version on SSL connections to this backend.",
		},
		"override_host": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Hostname to override the Host header.",
		},
		"prefer_ipv6": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Prefer IPv6 connections to origins for hostname backends. Default `false` for CDN services.",
		},
		"request_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Name of a request condition which, if met, selects this backend.",
		},
		"share_key": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Value that, when shared across backends, enables those backends to share the same health check.",
		},
		"shield": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "POP of the shield designated to reduce inbound load.",
		},
		"ssl_ca_cert": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "CA certificate attached to origin.",
		},
		"ssl_cert_hostname": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Hostname used for certificate validation. Does not affect SNI.",
		},
		"ssl_check_cert": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Whether to strictly check SSL certificates. Default `true`.",
		},
		"ssl_ciphers": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Cipher list for TLS connections to this backend.",
		},
		"ssl_client_cert": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Default:     stringdefault.StaticString(""),
			Description: "Client certificate used when connecting to the backend.",
		},
		"ssl_client_key": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Default:     stringdefault.StaticString(""),
			Description: "Client key used when connecting to the backend.",
		},
		"ssl_sni_hostname": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Hostname used for SNI in the TLS handshake.",
		},
		"use_ssl": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Whether to use SSL to reach the backend. Default `false`.",
		},
		"weight": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(100),
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
			if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.StatusCode == 404 {
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
		if !ModelsEqual(desiredBackend, remoteModel) {
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
		if !ModelsEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

func ModelsEqual(a, b NestedModel) bool {
	return service.StringValue(a.Name) == service.StringValue(b.Name) &&
		service.StringValue(a.Address) == service.StringValue(b.Address) &&
		service.Int64Value(a.Port) == service.Int64Value(b.Port) &&
		service.StringValue(a.Comment) == service.StringValue(b.Comment) &&
		service.BoolValue(a.AutoLoadbalance) == service.BoolValue(b.AutoLoadbalance) &&
		service.Int64Value(a.BetweenBytesTimeout) == service.Int64Value(b.BetweenBytesTimeout) &&
		service.Int64Value(a.ConnectTimeout) == service.Int64Value(b.ConnectTimeout) &&
		service.Int64Value(a.ErrorThreshold) == service.Int64Value(b.ErrorThreshold) &&
		service.Int64Value(a.FirstByteTimeout) == service.Int64Value(b.FirstByteTimeout) &&
		service.StringValue(a.HealthCheck) == service.StringValue(b.HealthCheck) &&
		service.Int64Value(a.KeepaliveTime) == service.Int64Value(b.KeepaliveTime) &&
		service.Int64Value(a.MaxConn) == service.Int64Value(b.MaxConn) &&
		service.Int64Value(a.MaxLifetime) == service.Int64Value(b.MaxLifetime) &&
		service.StringValue(a.MaxTLSVersion) == service.StringValue(b.MaxTLSVersion) &&
		service.Int64Value(a.MaxUse) == service.Int64Value(b.MaxUse) &&
		service.StringValue(a.MinTLSVersion) == service.StringValue(b.MinTLSVersion) &&
		service.StringValue(a.OverrideHost) == service.StringValue(b.OverrideHost) &&
		service.BoolValue(a.PreferIPv6) == service.BoolValue(b.PreferIPv6) &&
		service.StringValue(a.RequestCondition) == service.StringValue(b.RequestCondition) &&
		service.StringValue(a.ShareKey) == service.StringValue(b.ShareKey) &&
		service.StringValue(a.Shield) == service.StringValue(b.Shield) &&
		service.StringValue(a.SSLCACert) == service.StringValue(b.SSLCACert) &&
		service.StringValue(a.SSLCertHostname) == service.StringValue(b.SSLCertHostname) &&
		service.BoolValue(a.SSLCheckCert) == service.BoolValue(b.SSLCheckCert) &&
		service.StringValue(a.SSLCiphers) == service.StringValue(b.SSLCiphers) &&
		service.StringValue(a.SSLClientCert) == service.StringValue(b.SSLClientCert) &&
		service.StringValue(a.SSLClientKey) == service.StringValue(b.SSLClientKey) &&
		service.StringValue(a.SSLSNIHostname) == service.StringValue(b.SSLSNIHostname) &&
		service.BoolValue(a.UseSSL) == service.BoolValue(b.UseSSL) &&
		service.Int64Value(a.Weight) == service.Int64Value(b.Weight)
}
