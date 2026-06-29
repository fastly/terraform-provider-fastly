package backend

import (
	fastly "github.com/fastly/go-fastly/v15/fastly"

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

func BuildUpdateInput(serviceID string, version int, plan NestedModel) *fastly.UpdateBackendInput {
	input := &fastly.UpdateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(plan.Name),
	}

	input.Address = new(service.StringValue(plan.Address))
	input.Port = new(int(service.Int64Value(plan.Port)))
	input.BetweenBytesTimeout = new(int(service.Int64Value(plan.BetweenBytesTimeout)))
	input.ConnectTimeout = new(int(service.Int64Value(plan.ConnectTimeout)))
	input.ErrorThreshold = new(int(service.Int64Value(plan.ErrorThreshold)))
	input.FirstByteTimeout = new(int(service.Int64Value(plan.FirstByteTimeout)))
	input.HealthCheck = new(service.StringValue(plan.HealthCheck))
	input.MaxConn = new(int(service.Int64Value(plan.MaxConn)))
	input.Weight = new(int(service.Int64Value(plan.Weight)))
	input.PreferIPv6 = new(fastly.Compatibool(service.BoolValue(plan.PreferIPv6)))
	input.SSLCheckCert = new(fastly.Compatibool(service.BoolValue(plan.SSLCheckCert)))
	input.Shield = new(service.StringValue(plan.Shield))
	input.UseSSL = new(fastly.Compatibool(service.BoolValue(plan.UseSSL)))
	input.AutoLoadbalance = new(fastly.Compatibool(service.BoolValue(plan.AutoLoadbalance)))
	input.RequestCondition = new(service.StringValue(plan.RequestCondition))

	input.KeepAliveTime = fastly.NullInt(int(service.Int64Value(plan.KeepaliveTime)))
	input.MaxLifetime = fastly.NullInt(int(service.Int64Value(plan.MaxLifetime)))
	input.MaxUse = fastly.NullInt(int(service.Int64Value(plan.MaxUse)))
	input.Comment = fastly.NullString(service.StringValue(plan.Comment))
	input.MinTLSVersion = fastly.NullString(service.StringValue(plan.MinTLSVersion))
	input.MaxTLSVersion = fastly.NullString(service.StringValue(plan.MaxTLSVersion))
	input.OverrideHost = fastly.NullString(service.StringValue(plan.OverrideHost))
	input.ShareKey = fastly.NullString(service.StringValue(plan.ShareKey))
	input.SSLCACert = fastly.NullString(service.StringValue(plan.SSLCACert))
	input.SSLCertHostname = fastly.NullString(service.StringValue(plan.SSLCertHostname))
	input.SSLCiphers = fastly.NullString(service.StringValue(plan.SSLCiphers))
	input.SSLClientCert = fastly.NullString(service.StringValue(plan.SSLClientCert()))
	input.SSLClientKey = fastly.NullString(service.StringValue(plan.SSLClientKey()))
	input.SSLSNIHostname = fastly.NullString(service.StringValue(plan.SSLSNIHostname))

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
	input.SSLClientCert = fastly.NullString(service.StringValue(m.SSLClientCert()))
	input.SSLClientKey = fastly.NullString(service.StringValue(m.SSLClientKey()))
	input.SSLSNIHostname = fastly.NullString(service.StringValue(m.SSLSNIHostname))
}
