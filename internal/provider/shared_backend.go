package provider

import (
	"context"
	"sort"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceBackendNestedModel struct {
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

func backendCommonAttributes() map[string]schema.Attribute {
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

func backendResourceAttributes() map[string]schema.Attribute {
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
	for k, v := range backendCommonAttributes() {
		attrs[k] = v
	}
	return attrs
}

func backendNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Backends attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: backendCommonAttributes(),
		},
	}
}

func normalizeBackendModels(input []serviceBackendNestedModel) []serviceBackendNestedModel {
	out := make([]serviceBackendNestedModel, len(input))
	copy(out, input)

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name.ValueString() < out[j].Name.ValueString()
	})

	return out
}

func readBackendsForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]serviceBackendNestedModel, error) {
	remote, err := client.ListBackends(ctx, &fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	result := make([]serviceBackendNestedModel, 0, len(remote))
	for _, b := range remote {
		result = append(result, flattenBackendToVCLModel(b))
	}

	return normalizeBackendModels(result), nil
}

func reconcileBackends(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []serviceBackendNestedModel) error {
	remote, err := client.ListBackends(ctx, &fastly.ListBackendsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}

	desired = normalizeBackendModels(desired)

	remoteByName := make(map[string]*fastly.Backend, len(remote))
	for _, b := range remote {
		remoteByName[fastly.ToValue(b.Name)] = b
	}

	desiredByName := make(map[string]serviceBackendNestedModel, len(desired))
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
			input := buildCreateBackendInput(serviceID, version, desiredBackend)
			if _, err := client.CreateBackend(ctx, input); err != nil {
				return err
			}
			continue
		}

		remoteModel := flattenBackendToVCLModel(remoteBackend)
		if !backendModelsEqual(desiredBackend, remoteModel) {
			input := buildUpdateBackendInput(serviceID, version, desiredBackend, &remoteModel, false)
			if _, err := client.UpdateBackend(ctx, input); err != nil {
				return err
			}
		}
	}

	return nil
}

func backendsEqual(a, b []serviceBackendNestedModel) bool {
	a = normalizeBackendModels(a)
	b = normalizeBackendModels(b)

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !backendModelsEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

func backendModelsEqual(a, b serviceBackendNestedModel) bool {
	return stringValue(a.Name) == stringValue(b.Name) &&
		stringValue(a.Address) == stringValue(b.Address) &&
		int64Value(a.Port) == int64Value(b.Port) &&
		stringValue(a.Comment) == stringValue(b.Comment) &&
		boolValue(a.AutoLoadbalance) == boolValue(b.AutoLoadbalance) &&
		int64Value(a.BetweenBytesTimeout) == int64Value(b.BetweenBytesTimeout) &&
		int64Value(a.ConnectTimeout) == int64Value(b.ConnectTimeout) &&
		int64Value(a.ErrorThreshold) == int64Value(b.ErrorThreshold) &&
		int64Value(a.FirstByteTimeout) == int64Value(b.FirstByteTimeout) &&
		stringValue(a.HealthCheck) == stringValue(b.HealthCheck) &&
		int64Value(a.KeepaliveTime) == int64Value(b.KeepaliveTime) &&
		int64Value(a.MaxConn) == int64Value(b.MaxConn) &&
		int64Value(a.MaxLifetime) == int64Value(b.MaxLifetime) &&
		stringValue(a.MaxTLSVersion) == stringValue(b.MaxTLSVersion) &&
		int64Value(a.MaxUse) == int64Value(b.MaxUse) &&
		stringValue(a.MinTLSVersion) == stringValue(b.MinTLSVersion) &&
		stringValue(a.OverrideHost) == stringValue(b.OverrideHost) &&
		boolValue(a.PreferIPv6) == boolValue(b.PreferIPv6) &&
		stringValue(a.RequestCondition) == stringValue(b.RequestCondition) &&
		stringValue(a.ShareKey) == stringValue(b.ShareKey) &&
		stringValue(a.Shield) == stringValue(b.Shield) &&
		stringValue(a.SSLCACert) == stringValue(b.SSLCACert) &&
		stringValue(a.SSLCertHostname) == stringValue(b.SSLCertHostname) &&
		boolValue(a.SSLCheckCert) == boolValue(b.SSLCheckCert) &&
		stringValue(a.SSLCiphers) == stringValue(b.SSLCiphers) &&
		stringValue(a.SSLClientCert) == stringValue(b.SSLClientCert) &&
		stringValue(a.SSLClientKey) == stringValue(b.SSLClientKey) &&
		stringValue(a.SSLSNIHostname) == stringValue(b.SSLSNIHostname) &&
		boolValue(a.UseSSL) == boolValue(b.UseSSL) &&
		int64Value(a.Weight) == int64Value(b.Weight)
}

func buildCreateBackendInput(serviceID string, version int, m serviceBackendNestedModel) *fastly.CreateBackendInput {
	input := &fastly.CreateBackendInput{
		ServiceID:           serviceID,
		ServiceVersion:      version,
		Name:                fastly.ToPointer(stringValue(m.Name)),
		Address:             fastly.ToPointer(stringValue(m.Address)),
		Port:                fastly.ToPointer(int(int64Value(m.Port))),
		BetweenBytesTimeout: fastly.ToPointer(int(int64Value(m.BetweenBytesTimeout))),
		ConnectTimeout:      fastly.ToPointer(int(int64Value(m.ConnectTimeout))),
		ErrorThreshold:      fastly.ToPointer(int(int64Value(m.ErrorThreshold))),
		FirstByteTimeout:    fastly.ToPointer(int(int64Value(m.FirstByteTimeout))),
		HealthCheck:         fastly.ToPointer(stringValue(m.HealthCheck)),
		MaxConn:             fastly.ToPointer(int(int64Value(m.MaxConn))),
		PreferIPv6:          fastly.ToPointer(fastly.Compatibool(boolValue(m.PreferIPv6))),
		SSLCheckCert:        fastly.ToPointer(fastly.Compatibool(boolValue(m.SSLCheckCert))),
		Shield:              fastly.ToPointer(stringValue(m.Shield)),
		UseSSL:              fastly.ToPointer(fastly.Compatibool(boolValue(m.UseSSL))),
		Weight:              fastly.ToPointer(int(int64Value(m.Weight))),
		AutoLoadbalance:     fastly.ToPointer(fastly.Compatibool(boolValue(m.AutoLoadbalance))),
		RequestCondition:    fastly.ToPointer(stringValue(m.RequestCondition)),
	}

	if int64Value(m.KeepaliveTime) > 0 {
		input.KeepAliveTime = fastly.ToPointer(int(int64Value(m.KeepaliveTime)))
	}
	if int64Value(m.MaxLifetime) > 0 {
		input.MaxLifetime = fastly.ToPointer(int(int64Value(m.MaxLifetime)))
	}
	if int64Value(m.MaxUse) > 0 {
		input.MaxUse = fastly.ToPointer(int(int64Value(m.MaxUse)))
	}
	if stringValue(m.Comment) != "" {
		input.Comment = fastly.ToPointer(stringValue(m.Comment))
	}
	setCreateOnlyNonEmptyBackendStrings(input, m)

	return input
}

func buildUpdateBackendInput(serviceID string, version int, plan serviceBackendNestedModel, state *serviceBackendNestedModel, forceAll bool) *fastly.UpdateBackendInput {
	input := &fastly.UpdateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           stringValue(plan.Name),
	}

	setString := func(attr types.String, old types.String, setter func(string)) {
		if forceAll || state == nil || stringValue(attr) != stringValue(old) {
			setter(stringValue(attr))
		}
	}
	setInt := func(attr types.Int64, old types.Int64, setter func(int)) {
		if forceAll || state == nil || int64Value(attr) != int64Value(old) {
			setter(int(int64Value(attr)))
		}
	}
	setBool := func(attr types.Bool, old types.Bool, setter func(bool)) {
		if forceAll || state == nil || boolValue(attr) != boolValue(old) {
			setter(boolValue(attr))
		}
	}

	var old serviceBackendNestedModel
	if state != nil {
		old = *state
	}

	setString(plan.Address, old.Address, func(v string) { input.Address = fastly.ToPointer(v) })
	setInt(plan.Port, old.Port, func(v int) { input.Port = fastly.ToPointer(v) })
	setInt(plan.BetweenBytesTimeout, old.BetweenBytesTimeout, func(v int) { input.BetweenBytesTimeout = fastly.ToPointer(v) })
	setInt(plan.ConnectTimeout, old.ConnectTimeout, func(v int) { input.ConnectTimeout = fastly.ToPointer(v) })
	setInt(plan.ErrorThreshold, old.ErrorThreshold, func(v int) { input.ErrorThreshold = fastly.ToPointer(v) })
	setInt(plan.FirstByteTimeout, old.FirstByteTimeout, func(v int) { input.FirstByteTimeout = fastly.ToPointer(v) })
	setString(plan.HealthCheck, old.HealthCheck, func(v string) { input.HealthCheck = fastly.ToPointer(v) })
	setInt(plan.KeepaliveTime, old.KeepaliveTime, func(v int) { input.KeepAliveTime = fastly.ToPointer(v) })
	setInt(plan.MaxConn, old.MaxConn, func(v int) { input.MaxConn = fastly.ToPointer(v) })
	setInt(plan.MaxLifetime, old.MaxLifetime, func(v int) { input.MaxLifetime = fastly.ToPointer(v) })
	setString(plan.MaxTLSVersion, old.MaxTLSVersion, func(v string) { input.MaxTLSVersion = fastly.ToPointer(v) })
	setInt(plan.MaxUse, old.MaxUse, func(v int) { input.MaxUse = fastly.ToPointer(v) })
	setString(plan.MinTLSVersion, old.MinTLSVersion, func(v string) { input.MinTLSVersion = fastly.ToPointer(v) })
	setString(plan.OverrideHost, old.OverrideHost, func(v string) { input.OverrideHost = fastly.ToPointer(v) })
	setBool(plan.PreferIPv6, old.PreferIPv6, func(v bool) { input.PreferIPv6 = fastly.ToPointer(fastly.Compatibool(v)) })
	setString(plan.RequestCondition, old.RequestCondition, func(v string) { input.RequestCondition = fastly.ToPointer(v) })
	setString(plan.ShareKey, old.ShareKey, func(v string) { input.ShareKey = fastly.ToPointer(v) })
	setString(plan.Shield, old.Shield, func(v string) { input.Shield = fastly.ToPointer(v) })
	setString(plan.SSLCACert, old.SSLCACert, func(v string) { input.SSLCACert = fastly.ToPointer(v) })
	setString(plan.SSLCertHostname, old.SSLCertHostname, func(v string) { input.SSLCertHostname = fastly.ToPointer(v) })
	setBool(plan.SSLCheckCert, old.SSLCheckCert, func(v bool) { input.SSLCheckCert = fastly.ToPointer(fastly.Compatibool(v)) })
	setString(plan.SSLCiphers, old.SSLCiphers, func(v string) { input.SSLCiphers = fastly.ToPointer(v) })
	setString(plan.SSLClientCert, old.SSLClientCert, func(v string) { input.SSLClientCert = fastly.ToPointer(v) })
	setString(plan.SSLClientKey, old.SSLClientKey, func(v string) { input.SSLClientKey = fastly.ToPointer(v) })
	setString(plan.SSLSNIHostname, old.SSLSNIHostname, func(v string) { input.SSLSNIHostname = fastly.ToPointer(v) })
	setBool(plan.UseSSL, old.UseSSL, func(v bool) { input.UseSSL = fastly.ToPointer(fastly.Compatibool(v)) })
	setInt(plan.Weight, old.Weight, func(v int) { input.Weight = fastly.ToPointer(v) })
	setBool(plan.AutoLoadbalance, old.AutoLoadbalance, func(v bool) { input.AutoLoadbalance = fastly.ToPointer(fastly.Compatibool(v)) })
	setString(plan.Comment, old.Comment, func(v string) { input.Comment = fastly.ToPointer(v) })

	return input
}

func setCreateOnlyNonEmptyBackendStrings(input *fastly.CreateBackendInput, m serviceBackendNestedModel) {
	if stringValue(m.MinTLSVersion) != "" {
		input.MinTLSVersion = fastly.ToPointer(stringValue(m.MinTLSVersion))
	}
	if stringValue(m.MaxTLSVersion) != "" {
		input.MaxTLSVersion = fastly.ToPointer(stringValue(m.MaxTLSVersion))
	}
	if stringValue(m.OverrideHost) != "" {
		input.OverrideHost = fastly.ToPointer(stringValue(m.OverrideHost))
	}
	if stringValue(m.ShareKey) != "" {
		input.ShareKey = fastly.ToPointer(stringValue(m.ShareKey))
	}
	if stringValue(m.SSLCACert) != "" {
		input.SSLCACert = fastly.ToPointer(stringValue(m.SSLCACert))
	}
	if stringValue(m.SSLCertHostname) != "" {
		input.SSLCertHostname = fastly.ToPointer(stringValue(m.SSLCertHostname))
	}
	if stringValue(m.SSLCiphers) != "" {
		input.SSLCiphers = fastly.ToPointer(stringValue(m.SSLCiphers))
	}
	if stringValue(m.SSLClientCert) != "" {
		input.SSLClientCert = fastly.ToPointer(stringValue(m.SSLClientCert))
	}
	if stringValue(m.SSLClientKey) != "" {
		input.SSLClientKey = fastly.ToPointer(stringValue(m.SSLClientKey))
	}
	if stringValue(m.SSLSNIHostname) != "" {
		input.SSLSNIHostname = fastly.ToPointer(stringValue(m.SSLSNIHostname))
	}
}

func flattenBackendToVCLModel(b *fastly.Backend) serviceBackendNestedModel {
	m := serviceBackendNestedModel{}

	if b == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(b.Name))
	m.Address = types.StringValue(fastly.ToValue(b.Address))
	m.Port = int64PointerOrDefault(b.Port, 80)
	m.AutoLoadbalance = boolPointerOrDefault(b.AutoLoadbalance, false)
	m.BetweenBytesTimeout = int64PointerOrDefault(b.BetweenBytesTimeout, 10000)
	m.ConnectTimeout = int64PointerOrDefault(b.ConnectTimeout, 1000)
	m.ErrorThreshold = int64PointerOrDefault(b.ErrorThreshold, 0)
	m.FirstByteTimeout = int64PointerOrDefault(b.FirstByteTimeout, 15000)
	m.HealthCheck = stringPointerOrDefault(b.HealthCheck, "")
	m.KeepaliveTime = int64PointerOrNull(b.KeepAliveTime)
	m.MaxConn = int64PointerOrDefault(b.MaxConn, 200)
	m.MaxLifetime = int64PointerOrDefault(b.MaxLifetime, 0)
	m.MaxTLSVersion = stringPointerOrDefault(b.MaxTLSVersion, "")
	m.MaxUse = int64PointerOrDefault(b.MaxUse, 0)
	m.MinTLSVersion = stringPointerOrDefault(b.MinTLSVersion, "")
	m.OverrideHost = stringPointerOrDefault(b.OverrideHost, "")
	m.PreferIPv6 = boolPointerOrDefault(b.PreferIPv6, false)
	m.RequestCondition = stringPointerOrDefault(b.RequestCondition, "")
	m.ShareKey = stringPointerOrDefault(b.ShareKey, "")
	m.Shield = stringPointerOrDefault(b.Shield, "")
	m.SSLCACert = stringPointerOrDefault(b.SSLCACert, "")
	m.SSLCertHostname = stringPointerOrDefault(b.SSLCertHostname, "")
	m.SSLCheckCert = boolPointerOrDefault(b.SSLCheckCert, true)
	m.SSLCiphers = stringPointerOrDefault(b.SSLCiphers, "")
	m.SSLClientCert = stringPointerOrDefault(b.SSLClientCert, "")
	m.SSLClientKey = stringPointerOrDefault(b.SSLClientKey, "")
	m.SSLSNIHostname = stringPointerOrDefault(b.SSLSNIHostname, "")
	m.UseSSL = boolPointerOrDefault(b.UseSSL, false)
	m.Weight = int64PointerOrDefault(b.Weight, 100)

	if b.Comment != nil && *b.Comment != "" {
		m.Comment = types.StringValue(*b.Comment)
	} else {
		m.Comment = types.StringNull()
	}

	return m
}

func explicitBackendToVCLModel(m serviceBackendModel) serviceBackendNestedModel {
	return serviceBackendNestedModel{
		Name:                m.Name,
		Address:             m.Address,
		Port:                m.Port,
		Comment:             m.Comment,
		AutoLoadbalance:     m.AutoLoadbalance,
		BetweenBytesTimeout: m.BetweenBytesTimeout,
		ConnectTimeout:      m.ConnectTimeout,
		ErrorThreshold:      m.ErrorThreshold,
		FirstByteTimeout:    m.FirstByteTimeout,
		HealthCheck:         m.HealthCheck,
		KeepaliveTime:       m.KeepaliveTime,
		MaxConn:             m.MaxConn,
		MaxLifetime:         m.MaxLifetime,
		MaxTLSVersion:       m.MaxTLSVersion,
		MaxUse:              m.MaxUse,
		MinTLSVersion:       m.MinTLSVersion,
		OverrideHost:        m.OverrideHost,
		PreferIPv6:          m.PreferIPv6,
		RequestCondition:    m.RequestCondition,
		ShareKey:            m.ShareKey,
		Shield:              m.Shield,
		SSLCACert:           m.SSLCACert,
		SSLCertHostname:     m.SSLCertHostname,
		SSLCheckCert:        m.SSLCheckCert,
		SSLCiphers:          m.SSLCiphers,
		SSLClientCert:       m.SSLClientCert,
		SSLClientKey:        m.SSLClientKey,
		SSLSNIHostname:      m.SSLSNIHostname,
		UseSSL:              m.UseSSL,
		Weight:              m.Weight,
	}
}

func applyVCLModelToExplicitBackend(src serviceBackendNestedModel, dst *serviceBackendModel) {
	dst.Name = src.Name
	dst.Address = src.Address
	dst.Port = src.Port
	dst.Comment = src.Comment
	dst.AutoLoadbalance = src.AutoLoadbalance
	dst.BetweenBytesTimeout = src.BetweenBytesTimeout
	dst.ConnectTimeout = src.ConnectTimeout
	dst.ErrorThreshold = src.ErrorThreshold
	dst.FirstByteTimeout = src.FirstByteTimeout
	dst.HealthCheck = src.HealthCheck
	dst.KeepaliveTime = src.KeepaliveTime
	dst.MaxConn = src.MaxConn
	dst.MaxLifetime = src.MaxLifetime
	dst.MaxTLSVersion = src.MaxTLSVersion
	dst.MaxUse = src.MaxUse
	dst.MinTLSVersion = src.MinTLSVersion
	dst.OverrideHost = src.OverrideHost
	dst.PreferIPv6 = src.PreferIPv6
	dst.RequestCondition = src.RequestCondition
	dst.ShareKey = src.ShareKey
	dst.Shield = src.Shield
	dst.SSLCACert = src.SSLCACert
	dst.SSLCertHostname = src.SSLCertHostname
	dst.SSLCheckCert = src.SSLCheckCert
	dst.SSLCiphers = src.SSLCiphers
	dst.SSLClientCert = src.SSLClientCert
	dst.SSLClientKey = src.SSLClientKey
	dst.SSLSNIHostname = src.SSLSNIHostname
	dst.UseSSL = src.UseSSL
	dst.Weight = src.Weight
}
