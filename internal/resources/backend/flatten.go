package backend

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(b *fastly.Backend) NestedModel {
	m := NestedModel{}

	if b == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(b.Name))
	m.Address = types.StringValue(fastly.ToValue(b.Address))
	m.Port = service.Int64PointerOrDefault(b.Port, 80)
	m.AutoLoadbalance = service.BoolPointerOrDefault(b.AutoLoadbalance, false)
	m.BetweenBytesTimeout = service.Int64PointerOrDefault(b.BetweenBytesTimeout, 10000)
	m.ConnectTimeout = service.Int64PointerOrDefault(b.ConnectTimeout, 1000)
	m.ErrorThreshold = service.Int64PointerOrDefault(b.ErrorThreshold, 0)
	m.FirstByteTimeout = service.Int64PointerOrDefault(b.FirstByteTimeout, 15000)
	m.HealthCheck = service.StringPointerOrDefault(b.HealthCheck, "")
	m.KeepaliveTime = service.Int64PointerOrNull(b.KeepAliveTime)
	m.MaxConn = service.Int64PointerOrDefault(b.MaxConn, 200)
	m.MaxLifetime = service.Int64PointerOrDefault(b.MaxLifetime, 0)
	m.MaxTLSVersion = service.StringPointerOrDefault(b.MaxTLSVersion, "")
	m.MaxUse = service.Int64PointerOrDefault(b.MaxUse, 0)
	m.MinTLSVersion = service.StringPointerOrDefault(b.MinTLSVersion, "")
	m.OverrideHost = service.StringPointerOrDefault(b.OverrideHost, "")
	m.PreferIPv6 = service.BoolPointerOrDefault(b.PreferIPv6, false)
	m.RequestCondition = service.StringPointerOrDefault(b.RequestCondition, "")
	m.ShareKey = service.StringPointerOrDefault(b.ShareKey, "")
	m.Shield = service.StringPointerOrDefault(b.Shield, "")
	m.SSLCACert = service.StringPointerOrDefault(b.SSLCACert, "")
	m.SSLCertHostname = service.StringPointerOrDefault(b.SSLCertHostname, "")
	m.SSLCheckCert = service.BoolPointerOrDefault(b.SSLCheckCert, true)
	m.SSLCiphers = service.StringPointerOrDefault(b.SSLCiphers, "")
	m.SSLClientCert = service.StringPointerOrDefault(b.SSLClientCert, "")
	m.SSLClientKey = service.StringPointerOrDefault(b.SSLClientKey, "")
	m.SSLSNIHostname = service.StringPointerOrDefault(b.SSLSNIHostname, "")
	m.UseSSL = service.BoolPointerOrDefault(b.UseSSL, false)
	m.Weight = service.Int64PointerOrDefault(b.Weight, 100)

	if b.Comment != nil && *b.Comment != "" {
		m.Comment = types.StringValue(*b.Comment)
	} else {
		m.Comment = types.StringNull()
	}

	return m
}

func ApplyNestedToModel(src NestedModel, dst *Model) {
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

func flatten(ctx context.Context, b *fastly.Backend, m *Model) {
	if b == nil {
		tflog.Warn(ctx, "flatten called with nil backend")
		return
	}

	id := fastly.ToValue(b.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(b.ServiceVersion)) + "-" + fastly.ToValue(b.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(b.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(b.ServiceVersion)))

	backendModel := FlattenToNestedModel(b)
	ApplyNestedToModel(backendModel, m)

	tflog.Debug(ctx, "Flattened service backend state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
