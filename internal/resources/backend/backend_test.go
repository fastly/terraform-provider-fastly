package backend

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// Test helpers

// defaultNestedModel returns a NestedModel with default values for all fields
func defaultNestedModel() NestedModel {
	return NestedModel{
		Name:                types.StringValue(""),
		Address:             types.StringValue(""),
		Port:                types.Int64Value(80),
		Comment:             types.StringNull(),
		AutoLoadbalance:     types.BoolValue(false),
		BetweenBytesTimeout: types.Int64Value(10000),
		ConnectTimeout:      types.Int64Value(1000),
		ErrorThreshold:      types.Int64Value(0),
		FirstByteTimeout:    types.Int64Value(15000),
		HealthCheck:         types.StringValue(""),
		KeepaliveTime:       types.Int64Value(0),
		MaxConn:             types.Int64Value(200),
		MaxLifetime:         types.Int64Value(0),
		MaxTLSVersion:       types.StringValue(""),
		MaxUse:              types.Int64Value(0),
		MinTLSVersion:       types.StringValue(""),
		OverrideHost:        types.StringValue(""),
		PreferIPv6:          types.BoolValue(false),
		RequestCondition:    types.StringValue(""),
		ShareKey:            types.StringValue(""),
		Shield:              types.StringValue(""),
		SSLCACert:           types.StringValue(""),
		SSLCertHostname:     types.StringValue(""),
		SSLCheckCert:        types.BoolValue(true),
		SSLCiphers:          types.StringValue(""),
		SSLClientCert:       types.StringValue(""),
		SSLClientKey:        types.StringValue(""),
		SSLSNIHostname:      types.StringValue(""),
		UseSSL:              types.BoolValue(false),
		Weight:              types.Int64Value(100),
	}
}

// fullNestedModel returns a NestedModel with all fields populated with non-default values
func fullNestedModel() NestedModel {
	return NestedModel{
		Name:                types.StringValue("test-backend"),
		Address:             types.StringValue("api.example.com"),
		Port:                types.Int64Value(443),
		Comment:             types.StringValue("Test backend comment"),
		AutoLoadbalance:     types.BoolValue(true),
		BetweenBytesTimeout: types.Int64Value(5000),
		ConnectTimeout:      types.Int64Value(2000),
		ErrorThreshold:      types.Int64Value(5),
		FirstByteTimeout:    types.Int64Value(10000),
		HealthCheck:         types.StringValue("health-check-1"),
		KeepaliveTime:       types.Int64Value(300),
		MaxConn:             types.Int64Value(100),
		MaxLifetime:         types.Int64Value(60000),
		MaxTLSVersion:       types.StringValue("1.3"),
		MaxUse:              types.Int64Value(50),
		MinTLSVersion:       types.StringValue("1.2"),
		OverrideHost:        types.StringValue("override.example.com"),
		PreferIPv6:          types.BoolValue(true),
		RequestCondition:    types.StringValue("request-condition-1"),
		ShareKey:            types.StringValue("share-key-1"),
		Shield:              types.StringValue("iad-va-us"),
		SSLCACert:           types.StringValue("ca-cert-content"),
		SSLCertHostname:     types.StringValue("cert.example.com"),
		SSLCheckCert:        types.BoolValue(false),
		SSLCiphers:          types.StringValue("ECDHE-RSA-AES128-GCM-SHA256"),
		SSLClientCert:       types.StringValue("client-cert-content"),
		SSLClientKey:        types.StringValue("client-key-content"),
		SSLSNIHostname:      types.StringValue("sni.example.com"),
		UseSSL:              types.BoolValue(true),
		Weight:              types.Int64Value(75),
	}
}

// minimalNestedModel returns a NestedModel with only required fields for BuildCreateInput
func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-backend")
	m.Address = types.StringValue("api.example.com")
	return m
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		backend  *fastly.Backend
		expected NestedModel
	}{
		{
			name:     "nil backend returns empty model",
			backend:  nil,
			expected: NestedModel{},
		},
		{
			name: "backend with all fields populated",
			backend: &fastly.Backend{
				Name:                fastly.ToPointer("test-backend"),
				Address:             fastly.ToPointer("api.example.com"),
				Port:                fastly.ToPointer(443),
				Comment:             fastly.ToPointer("Test backend comment"),
				AutoLoadbalance:     fastly.ToPointer(true),
				BetweenBytesTimeout: fastly.ToPointer(5000),
				ConnectTimeout:      fastly.ToPointer(2000),
				ErrorThreshold:      fastly.ToPointer(5),
				FirstByteTimeout:    fastly.ToPointer(10000),
				HealthCheck:         fastly.ToPointer("health-check-1"),
				KeepAliveTime:       fastly.ToPointer(300),
				MaxConn:             fastly.ToPointer(100),
				MaxLifetime:         fastly.ToPointer(60000),
				MaxTLSVersion:       fastly.ToPointer("1.3"),
				MaxUse:              fastly.ToPointer(50),
				MinTLSVersion:       fastly.ToPointer("1.2"),
				OverrideHost:        fastly.ToPointer("override.example.com"),
				PreferIPv6:          fastly.ToPointer(true),
				RequestCondition:    fastly.ToPointer("request-condition-1"),
				ShareKey:            fastly.ToPointer("share-key-1"),
				Shield:              fastly.ToPointer("iad-va-us"),
				SSLCACert:           fastly.ToPointer("ca-cert-content"),
				SSLCertHostname:     fastly.ToPointer("cert.example.com"),
				SSLCheckCert:        fastly.ToPointer(false),
				SSLCiphers:          fastly.ToPointer("ECDHE-RSA-AES128-GCM-SHA256"),
				SSLClientCert:       fastly.ToPointer("client-cert-content"),
				SSLClientKey:        fastly.ToPointer("client-key-content"),
				SSLSNIHostname:      fastly.ToPointer("sni.example.com"),
				UseSSL:              fastly.ToPointer(true),
				Weight:              fastly.ToPointer(75),
			},
			expected: fullNestedModel(),
		},
		{
			name: "backend with default values",
			backend: &fastly.Backend{
				Name:    fastly.ToPointer("default-backend"),
				Address: fastly.ToPointer("api.default.com"),
			},
			expected: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("default-backend")
				m.Address = types.StringValue("api.default.com")
				return m
			}(),
		},
		{
			name: "backend with empty comment is null",
			backend: &fastly.Backend{
				Name:    fastly.ToPointer("backend-empty-comment"),
				Address: fastly.ToPointer("api.test.com"),
				Comment: fastly.ToPointer(""),
			},
			expected: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("backend-empty-comment")
				m.Address = types.StringValue("api.test.com")
				return m
			}(),
		},
		{
			name: "backend with nil KeepaliveTime",
			backend: &fastly.Backend{
				Name:          fastly.ToPointer("backend-nil-keepalive"),
				Address:       fastly.ToPointer("api.test.com"),
				KeepAliveTime: nil,
			},
			expected: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("backend-nil-keepalive")
				m.Address = types.StringValue("api.test.com")
				return m
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.backend)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyNestedToModel(t *testing.T) {
	src := fullNestedModel()
	dst := &Model{
		ID:      types.StringValue("existing-id"),
		Service: types.StringValue("existing-service"),
		Version: types.Int64Value(1),
	}

	ApplyNestedToModel(src, dst)

	// Verify ID/Service/Version are preserved
	assert.Equal(t, types.StringValue("existing-id"), dst.ID)
	assert.Equal(t, types.StringValue("existing-service"), dst.Service)
	assert.Equal(t, types.Int64Value(1), dst.Version)

	// Verify all other fields are copied
	assert.Equal(t, src.Name, dst.Name)
	assert.Equal(t, src.Address, dst.Address)
	assert.Equal(t, src.Port, dst.Port)
	assert.Equal(t, src.Comment, dst.Comment)
	assert.Equal(t, src.AutoLoadbalance, dst.AutoLoadbalance)
	assert.Equal(t, src.BetweenBytesTimeout, dst.BetweenBytesTimeout)
	assert.Equal(t, src.ConnectTimeout, dst.ConnectTimeout)
	assert.Equal(t, src.ErrorThreshold, dst.ErrorThreshold)
	assert.Equal(t, src.FirstByteTimeout, dst.FirstByteTimeout)
	assert.Equal(t, src.HealthCheck, dst.HealthCheck)
	assert.Equal(t, src.KeepaliveTime, dst.KeepaliveTime)
	assert.Equal(t, src.MaxConn, dst.MaxConn)
	assert.Equal(t, src.MaxLifetime, dst.MaxLifetime)
	assert.Equal(t, src.MaxTLSVersion, dst.MaxTLSVersion)
	assert.Equal(t, src.MaxUse, dst.MaxUse)
	assert.Equal(t, src.MinTLSVersion, dst.MinTLSVersion)
	assert.Equal(t, src.OverrideHost, dst.OverrideHost)
	assert.Equal(t, src.PreferIPv6, dst.PreferIPv6)
	assert.Equal(t, src.RequestCondition, dst.RequestCondition)
	assert.Equal(t, src.ShareKey, dst.ShareKey)
	assert.Equal(t, src.Shield, dst.Shield)
	assert.Equal(t, src.SSLCACert, dst.SSLCACert)
	assert.Equal(t, src.SSLCertHostname, dst.SSLCertHostname)
	assert.Equal(t, src.SSLCheckCert, dst.SSLCheckCert)
	assert.Equal(t, src.SSLCiphers, dst.SSLCiphers)
	assert.Equal(t, src.SSLClientCert, dst.SSLClientCert)
	assert.Equal(t, src.SSLClientKey, dst.SSLClientKey)
	assert.Equal(t, src.SSLSNIHostname, dst.SSLSNIHostname)
	assert.Equal(t, src.UseSSL, dst.UseSSL)
	assert.Equal(t, src.Weight, dst.Weight)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		backend  *fastly.Backend
		validate func(t *testing.T, m *Model)
	}{
		{
			name:    "nil backend logs warning",
			backend: nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "backend with service metadata",
			backend: &fastly.Backend{
				ServiceID:      fastly.ToPointer("service-123"),
				ServiceVersion: fastly.ToPointer(5),
				Name:           fastly.ToPointer("test-backend"),
				Address:        fastly.ToPointer("api.example.com"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-backend"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-backend"), m.Name)
				assert.Equal(t, types.StringValue("api.example.com"), m.Address)
			},
		},
		{
			name: "backend with all fields",
			backend: &fastly.Backend{
				ServiceID:           fastly.ToPointer("service-456"),
				ServiceVersion:      fastly.ToPointer(10),
				Name:                fastly.ToPointer("full-backend"),
				Address:             fastly.ToPointer("api.full.com"),
				Port:                fastly.ToPointer(8443),
				Comment:             fastly.ToPointer("Full backend"),
				AutoLoadbalance:     fastly.ToPointer(true),
				BetweenBytesTimeout: fastly.ToPointer(3000),
				UseSSL:              fastly.ToPointer(true),
				Weight:              fastly.ToPointer(50),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-456-10-full-backend"), m.ID)
				assert.Equal(t, types.StringValue("service-456"), m.Service)
				assert.Equal(t, types.Int64Value(10), m.Version)
				assert.Equal(t, types.StringValue("full-backend"), m.Name)
				assert.Equal(t, types.StringValue("api.full.com"), m.Address)
				assert.Equal(t, types.Int64Value(8443), m.Port)
				assert.Equal(t, types.StringValue("Full backend"), m.Comment)
				assert.Equal(t, types.BoolValue(true), m.AutoLoadbalance)
				assert.Equal(t, types.Int64Value(3000), m.BetweenBytesTimeout)
				assert.Equal(t, types.BoolValue(true), m.UseSSL)
				assert.Equal(t, types.Int64Value(50), m.Weight)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.backend, m)
			tt.validate(t, m)
		})
	}
}

// Tests for expand.go

func TestBuildCreateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.CreateBackendInput)
	}{
		{
			name:      "minimal backend",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-backend", *input.Name)
				assert.Equal(t, "api.example.com", *input.Address)
				assert.Equal(t, 80, *input.Port)
				assert.Equal(t, 10000, *input.BetweenBytesTimeout)
				assert.Equal(t, 1000, *input.ConnectTimeout)
				assert.Equal(t, 0, *input.ErrorThreshold)
				assert.Equal(t, 15000, *input.FirstByteTimeout)
				assert.Equal(t, "", *input.HealthCheck)
				assert.Equal(t, 200, *input.MaxConn)
				assert.Equal(t, fastly.Compatibool(false), *input.PreferIPv6)
				assert.Equal(t, fastly.Compatibool(true), *input.SSLCheckCert)
				assert.Equal(t, "", *input.Shield)
				assert.Equal(t, fastly.Compatibool(false), *input.UseSSL)
				assert.Equal(t, 100, *input.Weight)
				assert.Equal(t, fastly.Compatibool(false), *input.AutoLoadbalance)
				assert.Equal(t, "", *input.RequestCondition)
				assert.Nil(t, input.KeepAliveTime)
				assert.Nil(t, input.MaxLifetime)
				assert.Nil(t, input.MaxUse)
				assert.Nil(t, input.Comment)
			},
		},
		{
			name:      "backend with all fields",
			serviceID: "service-456",
			version:   10,
			model: func() NestedModel {
				m := fullNestedModel()
				m.Name = types.StringValue("full-backend")
				m.Address = types.StringValue("api.full.com")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.Equal(t, "service-456", input.ServiceID)
				assert.Equal(t, 10, input.ServiceVersion)
				assert.Equal(t, "full-backend", *input.Name)
				assert.Equal(t, "api.full.com", *input.Address)
				assert.Equal(t, 443, *input.Port)
				assert.Equal(t, "Test backend comment", *input.Comment)
				assert.Equal(t, fastly.Compatibool(true), *input.AutoLoadbalance)
				assert.Equal(t, 5000, *input.BetweenBytesTimeout)
				assert.Equal(t, 2000, *input.ConnectTimeout)
				assert.Equal(t, 5, *input.ErrorThreshold)
				assert.Equal(t, 10000, *input.FirstByteTimeout)
				assert.Equal(t, "health-check-1", *input.HealthCheck)
				assert.Equal(t, 300, *input.KeepAliveTime)
				assert.Equal(t, 100, *input.MaxConn)
				assert.Equal(t, 60000, *input.MaxLifetime)
				assert.Equal(t, "1.3", *input.MaxTLSVersion)
				assert.Equal(t, 50, *input.MaxUse)
				assert.Equal(t, "1.2", *input.MinTLSVersion)
				assert.Equal(t, "override.example.com", *input.OverrideHost)
				assert.Equal(t, fastly.Compatibool(true), *input.PreferIPv6)
				assert.Equal(t, "request-condition-1", *input.RequestCondition)
				assert.Equal(t, "share-key-1", *input.ShareKey)
				assert.Equal(t, "iad-va-us", *input.Shield)
				assert.Equal(t, "ca-cert-content", *input.SSLCACert)
				assert.Equal(t, "cert.example.com", *input.SSLCertHostname)
				assert.Equal(t, fastly.Compatibool(false), *input.SSLCheckCert)
				assert.Equal(t, "ECDHE-RSA-AES128-GCM-SHA256", *input.SSLCiphers)
				assert.Equal(t, "client-cert-content", *input.SSLClientCert)
				assert.Equal(t, "client-key-content", *input.SSLClientKey)
				assert.Equal(t, "sni.example.com", *input.SSLSNIHostname)
				assert.Equal(t, fastly.Compatibool(true), *input.UseSSL)
				assert.Equal(t, 75, *input.Weight)
			},
		},
		{
			name:      "backend with zero values for optional ints",
			serviceID: "service-789",
			version:   1,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("zero-backend")
				m.Address = types.StringValue("api.zero.com")
				m.KeepaliveTime = types.Int64Value(0)
				m.MaxLifetime = types.Int64Value(0)
				m.MaxUse = types.Int64Value(0)
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.Nil(t, input.KeepAliveTime, "KeepaliveTime should be nil when 0")
				assert.Nil(t, input.MaxLifetime, "MaxLifetime should be nil when 0")
				assert.Nil(t, input.MaxUse, "MaxUse should be nil when 0")
			},
		},
		{
			name:      "backend with empty comment",
			serviceID: "service-abc",
			version:   2,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("no-comment-backend")
				m.Address = types.StringValue("api.nocomment.com")
				m.Comment = types.StringValue("")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.Nil(t, input.Comment, "Comment should be nil when empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildCreateInput(tt.serviceID, tt.version, tt.model)
			tt.validate(t, input)
		})
	}
}

func TestBuildUpdateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		plan      NestedModel
		state     *NestedModel
		forceAll  bool
		validate  func(t *testing.T, input *fastly.UpdateBackendInput)
	}{
		{
			name:      "update with forceAll true",
			serviceID: "service-123",
			version:   6,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("test-backend")
				m.Address = types.StringValue("api.updated.com")
				m.Port = types.Int64Value(443)
				m.UseSSL = types.BoolValue(true)
				m.BetweenBytesTimeout = types.Int64Value(5000)
				m.ConnectTimeout = types.Int64Value(2000)
				return m
			}(),
			state: func() *NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("test-backend")
				m.Address = types.StringValue("api.example.com")
				return &m
			}(),
			forceAll: true,
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 6, input.ServiceVersion)
				assert.Equal(t, "test-backend", input.Name)
				assert.NotNil(t, input.Address)
				assert.Equal(t, "api.updated.com", *input.Address)
				assert.NotNil(t, input.Port)
				assert.Equal(t, 443, *input.Port)
				assert.NotNil(t, input.UseSSL)
				assert.Equal(t, fastly.Compatibool(true), *input.UseSSL)
			},
		},
		{
			name:      "update only changed fields",
			serviceID: "service-456",
			version:   10,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("backend-2")
				m.Address = types.StringValue("api.changed.com")
				m.Port = types.Int64Value(8080)
				return m
			}(),
			state: func() *NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("backend-2")
				m.Address = types.StringValue("api.original.com")
				return &m
			}(),
			forceAll: false,
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.Equal(t, "service-456", input.ServiceID)
				assert.Equal(t, 10, input.ServiceVersion)
				assert.Equal(t, "backend-2", input.Name)

				// Changed fields should be present
				assert.NotNil(t, input.Address)
				assert.Equal(t, "api.changed.com", *input.Address)
				assert.NotNil(t, input.Port)
				assert.Equal(t, 8080, *input.Port)

				// Unchanged fields should be nil (not sent in update payload)
				assert.Nil(t, input.UseSSL)
				assert.Nil(t, input.Weight)
				assert.Nil(t, input.ConnectTimeout)
				assert.Nil(t, input.BetweenBytesTimeout)
				assert.Nil(t, input.MaxConn)
				assert.Nil(t, input.AutoLoadbalance)
				assert.Nil(t, input.SSLCheckCert)
			},
		},
		{
			name:      "update with nil state forces all fields",
			serviceID: "service-789",
			version:   1,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("new-backend")
				m.Address = types.StringValue("api.new.com")
				m.Port = types.Int64Value(443)
				return m
			}(),
			state:    nil,
			forceAll: false,
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.Equal(t, "service-789", input.ServiceID)
				assert.Equal(t, 1, input.ServiceVersion)
				assert.Equal(t, "new-backend", input.Name)
				assert.NotNil(t, input.Address)
				assert.NotNil(t, input.Port)
				assert.NotNil(t, input.UseSSL)
			},
		},
		{
			name:      "update changes comment",
			serviceID: "service-comment",
			version:   3,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("commented-backend")
				m.Address = types.StringValue("api.test.com")
				m.Comment = types.StringValue("New comment")
				return m
			}(),
			state: func() *NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("commented-backend")
				m.Address = types.StringValue("api.test.com")
				m.Comment = types.StringValue("Old comment")
				return &m
			}(),
			forceAll: false,
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.NotNil(t, input.Comment)
				assert.Equal(t, "New comment", *input.Comment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildUpdateInput(tt.serviceID, tt.version, tt.plan, tt.state, tt.forceAll)
			tt.validate(t, input)
		})
	}
}

func TestSetCreateOnlyNonEmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		model    NestedModel
		validate func(t *testing.T, input *fastly.CreateBackendInput)
	}{
		{
			name: "all SSL fields populated",
			model: NestedModel{
				MinTLSVersion:   types.StringValue("1.2"),
				MaxTLSVersion:   types.StringValue("1.3"),
				OverrideHost:    types.StringValue("override.example.com"),
				ShareKey:        types.StringValue("share-key-1"),
				SSLCACert:       types.StringValue("ca-cert-content"),
				SSLCertHostname: types.StringValue("cert.example.com"),
				SSLCiphers:      types.StringValue("ECDHE-RSA-AES128-GCM-SHA256"),
				SSLClientCert:   types.StringValue("client-cert-content"),
				SSLClientKey:    types.StringValue("client-key-content"),
				SSLSNIHostname:  types.StringValue("sni.example.com"),
			},
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.NotNil(t, input.MinTLSVersion)
				assert.Equal(t, "1.2", *input.MinTLSVersion)
				assert.NotNil(t, input.MaxTLSVersion)
				assert.Equal(t, "1.3", *input.MaxTLSVersion)
				assert.NotNil(t, input.OverrideHost)
				assert.Equal(t, "override.example.com", *input.OverrideHost)
				assert.NotNil(t, input.ShareKey)
				assert.Equal(t, "share-key-1", *input.ShareKey)
				assert.NotNil(t, input.SSLCACert)
				assert.Equal(t, "ca-cert-content", *input.SSLCACert)
				assert.NotNil(t, input.SSLCertHostname)
				assert.Equal(t, "cert.example.com", *input.SSLCertHostname)
				assert.NotNil(t, input.SSLCiphers)
				assert.Equal(t, "ECDHE-RSA-AES128-GCM-SHA256", *input.SSLCiphers)
				assert.NotNil(t, input.SSLClientCert)
				assert.Equal(t, "client-cert-content", *input.SSLClientCert)
				assert.NotNil(t, input.SSLClientKey)
				assert.Equal(t, "client-key-content", *input.SSLClientKey)
				assert.NotNil(t, input.SSLSNIHostname)
				assert.Equal(t, "sni.example.com", *input.SSLSNIHostname)
			},
		},
		{
			name:  "all SSL fields empty",
			model: defaultNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.Nil(t, input.MinTLSVersion)
				assert.Nil(t, input.MaxTLSVersion)
				assert.Nil(t, input.OverrideHost)
				assert.Nil(t, input.ShareKey)
				assert.Nil(t, input.SSLCACert)
				assert.Nil(t, input.SSLCertHostname)
				assert.Nil(t, input.SSLCiphers)
				assert.Nil(t, input.SSLClientCert)
				assert.Nil(t, input.SSLClientKey)
				assert.Nil(t, input.SSLSNIHostname)
			},
		},
		{
			name: "mixed empty and populated fields",
			model: NestedModel{
				MinTLSVersion:   types.StringValue("1.2"),
				MaxTLSVersion:   types.StringValue(""),
				OverrideHost:    types.StringValue("override.example.com"),
				ShareKey:        types.StringValue(""),
				SSLCACert:       types.StringValue("ca-cert-content"),
				SSLCertHostname: types.StringValue(""),
				SSLCiphers:      types.StringValue(""),
				SSLClientCert:   types.StringValue("client-cert-content"),
				SSLClientKey:    types.StringValue("client-key-content"),
				SSLSNIHostname:  types.StringValue(""),
			},
			validate: func(t *testing.T, input *fastly.CreateBackendInput) {
				assert.NotNil(t, input.MinTLSVersion)
				assert.Equal(t, "1.2", *input.MinTLSVersion)
				assert.Nil(t, input.MaxTLSVersion)
				assert.NotNil(t, input.OverrideHost)
				assert.Equal(t, "override.example.com", *input.OverrideHost)
				assert.Nil(t, input.ShareKey)
				assert.NotNil(t, input.SSLCACert)
				assert.Equal(t, "ca-cert-content", *input.SSLCACert)
				assert.Nil(t, input.SSLCertHostname)
				assert.Nil(t, input.SSLCiphers)
				assert.NotNil(t, input.SSLClientCert)
				assert.Equal(t, "client-cert-content", *input.SSLClientCert)
				assert.NotNil(t, input.SSLClientKey)
				assert.Equal(t, "client-key-content", *input.SSLClientKey)
				assert.Nil(t, input.SSLSNIHostname)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &fastly.CreateBackendInput{}
			setCreateOnlyNonEmptyStrings(input, tt.model)
			tt.validate(t, input)
		})
	}
}

func TestModelToNested(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		expected NestedModel
	}{
		{
			name: "all fields populated",
			model: func() Model {
				nested := fullNestedModel()
				return Model{
					ID:                  types.StringValue("service-123-5-test"),
					Service:             types.StringValue("service-123"),
					Version:             types.Int64Value(5),
					Name:                nested.Name,
					Address:             nested.Address,
					Port:                nested.Port,
					Comment:             nested.Comment,
					AutoLoadbalance:     nested.AutoLoadbalance,
					BetweenBytesTimeout: nested.BetweenBytesTimeout,
					ConnectTimeout:      nested.ConnectTimeout,
					ErrorThreshold:      nested.ErrorThreshold,
					FirstByteTimeout:    nested.FirstByteTimeout,
					HealthCheck:         nested.HealthCheck,
					KeepaliveTime:       nested.KeepaliveTime,
					MaxConn:             nested.MaxConn,
					MaxLifetime:         nested.MaxLifetime,
					MaxTLSVersion:       nested.MaxTLSVersion,
					MaxUse:              nested.MaxUse,
					MinTLSVersion:       nested.MinTLSVersion,
					OverrideHost:        nested.OverrideHost,
					PreferIPv6:          nested.PreferIPv6,
					RequestCondition:    nested.RequestCondition,
					ShareKey:            nested.ShareKey,
					Shield:              nested.Shield,
					SSLCACert:           nested.SSLCACert,
					SSLCertHostname:     nested.SSLCertHostname,
					SSLCheckCert:        nested.SSLCheckCert,
					SSLCiphers:          nested.SSLCiphers,
					SSLClientCert:       nested.SSLClientCert,
					SSLClientKey:        nested.SSLClientKey,
					SSLSNIHostname:      nested.SSLSNIHostname,
					UseSSL:              nested.UseSSL,
					Weight:              nested.Weight,
				}
			}(),
			expected: fullNestedModel(),
		},
		{
			name: "minimal fields only",
			model: Model{
				ID:      types.StringValue("service-456-1-minimal"),
				Service: types.StringValue("service-456"),
				Version: types.Int64Value(1),
				Name:    types.StringValue("minimal-backend"),
				Address: types.StringValue("api.minimal.com"),
				Port:    types.Int64Value(80),
			},
			expected: NestedModel{
				Name:    types.StringValue("minimal-backend"),
				Address: types.StringValue("api.minimal.com"),
				Port:    types.Int64Value(80),
			},
		},
		{
			name: "null and empty values",
			model: Model{
				ID:                  types.StringValue("service-789-2-empty"),
				Service:             types.StringValue("service-789"),
				Version:             types.Int64Value(2),
				Name:                types.StringValue("empty-backend"),
				Address:             types.StringValue("api.empty.com"),
				Port:                types.Int64Value(443),
				Comment:             types.StringNull(),
				KeepaliveTime:       types.Int64Value(0),
				MaxTLSVersion:       types.StringValue(""),
				MinTLSVersion:       types.StringValue(""),
				OverrideHost:        types.StringValue(""),
				AutoLoadbalance:     types.BoolValue(false),
				BetweenBytesTimeout: types.Int64Value(10000),
				ConnectTimeout:      types.Int64Value(1000),
				ErrorThreshold:      types.Int64Value(0),
				FirstByteTimeout:    types.Int64Value(15000),
				HealthCheck:         types.StringValue(""),
				MaxConn:             types.Int64Value(200),
				MaxLifetime:         types.Int64Value(0),
				MaxUse:              types.Int64Value(0),
				PreferIPv6:          types.BoolValue(false),
				RequestCondition:    types.StringValue(""),
				ShareKey:            types.StringValue(""),
				Shield:              types.StringValue(""),
				SSLCACert:           types.StringValue(""),
				SSLCertHostname:     types.StringValue(""),
				SSLCheckCert:        types.BoolValue(true),
				SSLCiphers:          types.StringValue(""),
				SSLClientCert:       types.StringValue(""),
				SSLClientKey:        types.StringValue(""),
				SSLSNIHostname:      types.StringValue(""),
				UseSSL:              types.BoolValue(false),
				Weight:              types.Int64Value(100),
			},
			expected: NestedModel{
				Name:                types.StringValue("empty-backend"),
				Address:             types.StringValue("api.empty.com"),
				Port:                types.Int64Value(443),
				Comment:             types.StringNull(),
				KeepaliveTime:       types.Int64Value(0),
				MaxTLSVersion:       types.StringValue(""),
				MinTLSVersion:       types.StringValue(""),
				OverrideHost:        types.StringValue(""),
				AutoLoadbalance:     types.BoolValue(false),
				BetweenBytesTimeout: types.Int64Value(10000),
				ConnectTimeout:      types.Int64Value(1000),
				ErrorThreshold:      types.Int64Value(0),
				FirstByteTimeout:    types.Int64Value(15000),
				HealthCheck:         types.StringValue(""),
				MaxConn:             types.Int64Value(200),
				MaxLifetime:         types.Int64Value(0),
				MaxUse:              types.Int64Value(0),
				PreferIPv6:          types.BoolValue(false),
				RequestCondition:    types.StringValue(""),
				ShareKey:            types.StringValue(""),
				Shield:              types.StringValue(""),
				SSLCACert:           types.StringValue(""),
				SSLCertHostname:     types.StringValue(""),
				SSLCheckCert:        types.BoolValue(true),
				SSLCiphers:          types.StringValue(""),
				SSLClientCert:       types.StringValue(""),
				SSLClientKey:        types.StringValue(""),
				SSLSNIHostname:      types.StringValue(""),
				UseSSL:              types.BoolValue(false),
				Weight:              types.Int64Value(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ModelToNested(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}
