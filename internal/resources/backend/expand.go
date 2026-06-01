package backend

import (
	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateBackendInput {
	input := &fastly.CreateBackendInput{
		ServiceID:           serviceID,
		ServiceVersion:      version,
		Name:                fastly.ToPointer(service.StringValue(m.Name)),
		Address:             fastly.ToPointer(service.StringValue(m.Address)),
		Port:                fastly.ToPointer(int(service.Int64Value(m.Port))),
		BetweenBytesTimeout: fastly.ToPointer(int(service.Int64Value(m.BetweenBytesTimeout))),
		ConnectTimeout:      fastly.ToPointer(int(service.Int64Value(m.ConnectTimeout))),
		ErrorThreshold:      fastly.ToPointer(int(service.Int64Value(m.ErrorThreshold))),
		FirstByteTimeout:    fastly.ToPointer(int(service.Int64Value(m.FirstByteTimeout))),
		HealthCheck:         fastly.ToPointer(service.StringValue(m.HealthCheck)),
		MaxConn:             fastly.ToPointer(int(service.Int64Value(m.MaxConn))),
		PreferIPv6:          fastly.ToPointer(fastly.Compatibool(service.BoolValue(m.PreferIPv6))),
		SSLCheckCert:        fastly.ToPointer(fastly.Compatibool(service.BoolValue(m.SSLCheckCert))),
		Shield:              fastly.ToPointer(service.StringValue(m.Shield)),
		UseSSL:              fastly.ToPointer(fastly.Compatibool(service.BoolValue(m.UseSSL))),
		Weight:              fastly.ToPointer(int(service.Int64Value(m.Weight))),
		AutoLoadbalance:     fastly.ToPointer(fastly.Compatibool(service.BoolValue(m.AutoLoadbalance))),
		RequestCondition:    fastly.ToPointer(service.StringValue(m.RequestCondition)),
	}

	if service.Int64Value(m.KeepaliveTime) > 0 {
		input.KeepAliveTime = fastly.ToPointer(int(service.Int64Value(m.KeepaliveTime)))
	}
	if service.Int64Value(m.MaxLifetime) > 0 {
		input.MaxLifetime = fastly.ToPointer(int(service.Int64Value(m.MaxLifetime)))
	}
	if service.Int64Value(m.MaxUse) > 0 {
		input.MaxUse = fastly.ToPointer(int(service.Int64Value(m.MaxUse)))
	}
	if service.StringValue(m.Comment) != "" {
		input.Comment = fastly.ToPointer(service.StringValue(m.Comment))
	}
	setCreateOnlyNonEmptyStrings(input, m)

	return input
}

func BuildUpdateInput(serviceID string, version int, plan NestedModel, state *NestedModel, forceAll bool) *fastly.UpdateBackendInput {
	input := &fastly.UpdateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(plan.Name),
	}

	setString := func(attr types.String, old types.String, setter func(string)) {
		if forceAll || state == nil || service.StringValue(attr) != service.StringValue(old) {
			setter(service.StringValue(attr))
		}
	}
	setInt := func(attr types.Int64, old types.Int64, setter func(int)) {
		if forceAll || state == nil || service.Int64Value(attr) != service.Int64Value(old) {
			setter(int(service.Int64Value(attr)))
		}
	}
	setBool := func(attr types.Bool, old types.Bool, setter func(bool)) {
		if forceAll || state == nil || service.BoolValue(attr) != service.BoolValue(old) {
			setter(service.BoolValue(attr))
		}
	}

	var old NestedModel
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

func setCreateOnlyNonEmptyStrings(input *fastly.CreateBackendInput, m NestedModel) {
	if service.StringValue(m.MinTLSVersion) != "" {
		input.MinTLSVersion = fastly.ToPointer(service.StringValue(m.MinTLSVersion))
	}
	if service.StringValue(m.MaxTLSVersion) != "" {
		input.MaxTLSVersion = fastly.ToPointer(service.StringValue(m.MaxTLSVersion))
	}
	if service.StringValue(m.OverrideHost) != "" {
		input.OverrideHost = fastly.ToPointer(service.StringValue(m.OverrideHost))
	}
	if service.StringValue(m.ShareKey) != "" {
		input.ShareKey = fastly.ToPointer(service.StringValue(m.ShareKey))
	}
	if service.StringValue(m.SSLCACert) != "" {
		input.SSLCACert = fastly.ToPointer(service.StringValue(m.SSLCACert))
	}
	if service.StringValue(m.SSLCertHostname) != "" {
		input.SSLCertHostname = fastly.ToPointer(service.StringValue(m.SSLCertHostname))
	}
	if service.StringValue(m.SSLCiphers) != "" {
		input.SSLCiphers = fastly.ToPointer(service.StringValue(m.SSLCiphers))
	}
	if service.StringValue(m.SSLClientCert) != "" {
		input.SSLClientCert = fastly.ToPointer(service.StringValue(m.SSLClientCert))
	}
	if service.StringValue(m.SSLClientKey) != "" {
		input.SSLClientKey = fastly.ToPointer(service.StringValue(m.SSLClientKey))
	}
	if service.StringValue(m.SSLSNIHostname) != "" {
		input.SSLSNIHostname = fastly.ToPointer(service.StringValue(m.SSLSNIHostname))
	}
}

func ModelToNested(m Model) NestedModel {
	return NestedModel{
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
