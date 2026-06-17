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
		Name:                new(service.StringValue(m.Name)),
		Address:             new(service.StringValue(m.Address)),
		Port:                new(int(service.Int64Value(m.Port))),
		BetweenBytesTimeout: new(int(service.Int64Value(m.BetweenBytesTimeout))),
		ConnectTimeout:      new(int(service.Int64Value(m.ConnectTimeout))),
		ErrorThreshold:      new(int(service.Int64Value(m.ErrorThreshold))),
		FirstByteTimeout:    new(int(service.Int64Value(m.FirstByteTimeout))),
		HealthCheck:         new(service.StringValue(m.HealthCheck)),
		MaxConn:             new(int(service.Int64Value(m.MaxConn))),
		PreferIPv6:          new(fastly.Compatibool(service.BoolValue(m.PreferIPv6))),
		SSLCheckCert:        new(fastly.Compatibool(service.BoolValue(m.SSLCheckCert))),
		Shield:              new(service.StringValue(m.Shield)),
		UseSSL:              new(fastly.Compatibool(service.BoolValue(m.UseSSL))),
		Weight:              new(int(service.Int64Value(m.Weight))),
		AutoLoadbalance:     new(fastly.Compatibool(service.BoolValue(m.AutoLoadbalance))),
		RequestCondition:    new(service.StringValue(m.RequestCondition)),
	}

	input.KeepAliveTime = fastly.NullInt(int(service.Int64Value(m.KeepaliveTime)))
	input.MaxLifetime = fastly.NullInt(int(service.Int64Value(m.MaxLifetime)))
	input.MaxUse = fastly.NullInt(int(service.Int64Value(m.MaxUse)))
	input.Comment = fastly.NullString(service.StringValue(m.Comment))
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

	setString(plan.Address, old.Address, func(v string) { input.Address = new(v) })
	setInt(plan.Port, old.Port, func(v int) { input.Port = new(v) })
	setInt(plan.BetweenBytesTimeout, old.BetweenBytesTimeout, func(v int) { input.BetweenBytesTimeout = new(v) })
	setInt(plan.ConnectTimeout, old.ConnectTimeout, func(v int) { input.ConnectTimeout = new(v) })
	setInt(plan.ErrorThreshold, old.ErrorThreshold, func(v int) { input.ErrorThreshold = new(v) })
	setInt(plan.FirstByteTimeout, old.FirstByteTimeout, func(v int) { input.FirstByteTimeout = new(v) })
	setString(plan.HealthCheck, old.HealthCheck, func(v string) { input.HealthCheck = new(v) })
	setInt(plan.KeepaliveTime, old.KeepaliveTime, func(v int) { input.KeepAliveTime = new(v) })
	setInt(plan.MaxConn, old.MaxConn, func(v int) { input.MaxConn = new(v) })
	setInt(plan.MaxLifetime, old.MaxLifetime, func(v int) { input.MaxLifetime = new(v) })
	setString(plan.MaxTLSVersion, old.MaxTLSVersion, func(v string) { input.MaxTLSVersion = new(v) })
	setInt(plan.MaxUse, old.MaxUse, func(v int) { input.MaxUse = new(v) })
	setString(plan.MinTLSVersion, old.MinTLSVersion, func(v string) { input.MinTLSVersion = new(v) })
	setString(plan.OverrideHost, old.OverrideHost, func(v string) { input.OverrideHost = new(v) })
	setBool(plan.PreferIPv6, old.PreferIPv6, func(v bool) { input.PreferIPv6 = new(fastly.Compatibool(v)) })
	setString(plan.RequestCondition, old.RequestCondition, func(v string) { input.RequestCondition = new(v) })
	setString(plan.ShareKey, old.ShareKey, func(v string) { input.ShareKey = new(v) })
	setString(plan.Shield, old.Shield, func(v string) { input.Shield = new(v) })
	setString(plan.SSLCACert, old.SSLCACert, func(v string) { input.SSLCACert = new(v) })
	setString(plan.SSLCertHostname, old.SSLCertHostname, func(v string) { input.SSLCertHostname = new(v) })
	setBool(plan.SSLCheckCert, old.SSLCheckCert, func(v bool) { input.SSLCheckCert = new(fastly.Compatibool(v)) })
	setString(plan.SSLCiphers, old.SSLCiphers, func(v string) { input.SSLCiphers = new(v) })
	setString(plan.SSLClientCert, old.SSLClientCert, func(v string) { input.SSLClientCert = new(v) })
	setString(plan.SSLClientKey, old.SSLClientKey, func(v string) { input.SSLClientKey = new(v) })
	setString(plan.SSLSNIHostname, old.SSLSNIHostname, func(v string) { input.SSLSNIHostname = new(v) })
	setBool(plan.UseSSL, old.UseSSL, func(v bool) { input.UseSSL = new(fastly.Compatibool(v)) })
	setInt(plan.Weight, old.Weight, func(v int) { input.Weight = new(v) })
	setBool(plan.AutoLoadbalance, old.AutoLoadbalance, func(v bool) { input.AutoLoadbalance = new(fastly.Compatibool(v)) })
	setString(plan.Comment, old.Comment, func(v string) { input.Comment = new(v) })

	return input
}

func setCreateOnlyNonEmptyStrings(input *fastly.CreateBackendInput, m NestedModel) {
	input.MinTLSVersion = fastly.NullString(service.StringValue(m.MinTLSVersion))
	input.MaxTLSVersion = fastly.NullString(service.StringValue(m.MaxTLSVersion))
	input.OverrideHost = fastly.NullString(service.StringValue(m.OverrideHost))
	input.ShareKey = fastly.NullString(service.StringValue(m.ShareKey))
	input.SSLCACert = fastly.NullString(service.StringValue(m.SSLCACert))
	input.SSLCertHostname = fastly.NullString(service.StringValue(m.SSLCertHostname))
	input.SSLCiphers = fastly.NullString(service.StringValue(m.SSLCiphers))
	input.SSLClientCert = fastly.NullString(service.StringValue(m.SSLClientCert))
	input.SSLClientKey = fastly.NullString(service.StringValue(m.SSLClientKey))
	input.SSLSNIHostname = fastly.NullString(service.StringValue(m.SSLSNIHostname))
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
