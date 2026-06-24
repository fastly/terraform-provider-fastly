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
		KeepaliveTime:       types.Int64Null(),
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
				Name:                new("test-backend"),
				Address:             new("api.example.com"),
				Port:                new(443),
				Comment:             new("Test backend comment"),
				AutoLoadbalance:     new(true),
				BetweenBytesTimeout: new(5000),
				ConnectTimeout:      new(2000),
				ErrorThreshold:      new(5),
				FirstByteTimeout:    new(10000),
				HealthCheck:         new("health-check-1"),
				KeepAliveTime:       new(300),
				MaxConn:             new(100),
				MaxLifetime:         new(60000),
				MaxTLSVersion:       new("1.3"),
				MaxUse:              new(50),
				MinTLSVersion:       new("1.2"),
				OverrideHost:        new("override.example.com"),
				PreferIPv6:          new(true),
				RequestCondition:    new("request-condition-1"),
				ShareKey:            new("share-key-1"),
				Shield:              new("iad-va-us"),
				SSLCACert:           new("ca-cert-content"),
				SSLCertHostname:     new("cert.example.com"),
				SSLCheckCert:        new(false),
				SSLCiphers:          new("ECDHE-RSA-AES128-GCM-SHA256"),
				SSLClientCert:       new("client-cert-content"),
				SSLClientKey:        new("client-key-content"),
				SSLSNIHostname:      new("sni.example.com"),
				UseSSL:              new(true),
				Weight:              new(75),
			},
			expected: fullNestedModel(),
		},
		{
			name: "backend with default values",
			backend: &fastly.Backend{
				Name:    new("default-backend"),
				Address: new("api.default.com"),
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
				Name:    new("backend-empty-comment"),
				Address: new("api.test.com"),
				Comment: new(""),
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
				Name:          new("backend-nil-keepalive"),
				Address:       new("api.test.com"),
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

func TestModelEmbedding(t *testing.T) {
	nested := fullNestedModel()
	m := Model{
		NestedModel: nested,
		ID:          types.StringValue("test-id"),
		Service:     types.StringValue("test-service"),
		Version:     types.Int64Value(1),
	}

	// Verify embedded fields are accessible
	assert.Equal(t, nested.Name, m.Name)
	assert.Equal(t, nested.Address, m.Address)
	assert.Equal(t, nested.Port, m.Port)

	// Verify Model-specific fields
	assert.Equal(t, types.StringValue("test-id"), m.ID)
	assert.Equal(t, types.StringValue("test-service"), m.Service)
	assert.Equal(t, types.Int64Value(1), m.Version)

	// Verify NestedModel can be extracted
	extracted := m.NestedModel
	assert.Equal(t, nested, extracted)
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
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-backend"),
				Address:        new("api.example.com"),
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
				ServiceID:           new("service-456"),
				ServiceVersion:      new(10),
				Name:                new("full-backend"),
				Address:             new("api.full.com"),
				Port:                new(8443),
				Comment:             new("Full backend"),
				AutoLoadbalance:     new(true),
				BetweenBytesTimeout: new(3000),
				UseSSL:              new(true),
				Weight:              new(50),
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
		validate  func(t *testing.T, input *fastly.UpdateBackendInput)
	}{
		{
			name:      "update with all fields",
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
			name:      "update with changed address",
			serviceID: "service-456",
			version:   10,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("backend-2")
				m.Address = types.StringValue("api.changed.com")
				m.Port = types.Int64Value(8080)
				return m
			}(),
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.Equal(t, "service-456", input.ServiceID)
				assert.Equal(t, 10, input.ServiceVersion)
				assert.Equal(t, "backend-2", input.Name)
				assert.NotNil(t, input.Address)
				assert.Equal(t, "api.changed.com", *input.Address)
				assert.NotNil(t, input.Port)
				assert.Equal(t, 8080, *input.Port)
				assert.NotNil(t, input.UseSSL)
				assert.NotNil(t, input.Weight)
			},
		},
		{
			name:      "update with new backend",
			serviceID: "service-789",
			version:   1,
			plan: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("new-backend")
				m.Address = types.StringValue("api.new.com")
				m.Port = types.Int64Value(443)
				return m
			}(),
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
			validate: func(t *testing.T, input *fastly.UpdateBackendInput) {
				assert.NotNil(t, input.Comment)
				assert.Equal(t, "New comment", *input.Comment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildUpdateInput(tt.serviceID, tt.version, tt.plan)
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
			model: Model{
				NestedModel: fullNestedModel(),
				ID:          types.StringValue("service-123-5-test"),
				Service:     types.StringValue("service-123"),
				Version:     types.Int64Value(5),
			},
			expected: fullNestedModel(),
		},
		{
			name: "minimal fields only",
			model: Model{
				NestedModel: NestedModel{
					Name:    types.StringValue("minimal-backend"),
					Address: types.StringValue("api.minimal.com"),
					Port:    types.Int64Value(80),
				},
				ID:      types.StringValue("service-456-1-minimal"),
				Service: types.StringValue("service-456"),
				Version: types.Int64Value(1),
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
				NestedModel: NestedModel{
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
				ID:      types.StringValue("service-789-2-empty"),
				Service: types.StringValue("service-789"),
				Version: types.Int64Value(2),
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
			result := tt.model.NestedModel
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for schema.go

func TestModelsEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        NestedModel
		b        NestedModel
		expected bool
	}{
		{
			name:     "identical models",
			a:        fullNestedModel(),
			b:        fullNestedModel(),
			expected: true,
		},
		{
			name:     "default models",
			a:        defaultNestedModel(),
			b:        defaultNestedModel(),
			expected: true,
		},
		{
			name: "different name",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("backend-1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("backend-2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different address",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Address = types.StringValue("api1.example.com")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Address = types.StringValue("api2.example.com")
				return m
			}(),
			expected: false,
		},
		{
			name: "different port",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Port = types.Int64Value(80)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Port = types.Int64Value(443)
				return m
			}(),
			expected: false,
		},
		{
			name: "different ssl settings",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.UseSSL = types.BoolValue(true)
				m.SSLCheckCert = types.BoolValue(false)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.UseSSL = types.BoolValue(true)
				m.SSLCheckCert = types.BoolValue(true)
				return m
			}(),
			expected: false,
		},
		{
			name: "null vs empty string comment",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Comment = types.StringNull()
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Comment = types.StringValue("")
				return m
			}(),
			expected: true,
		},
		{
			name: "different weight",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Weight = types.Int64Value(100)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Weight = types.Int64Value(50)
				return m
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.ModelsEqual(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    []NestedModel
		expected []NestedModel
	}{
		{
			name:     "empty slice",
			input:    []NestedModel{},
			expected: []NestedModel{},
		},
		{
			name: "single backend",
			input: []NestedModel{
				{Name: types.StringValue("backend-1")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("backend-1")},
			},
		},
		{
			name: "already sorted",
			input: []NestedModel{
				{Name: types.StringValue("backend-a")},
				{Name: types.StringValue("backend-b")},
				{Name: types.StringValue("backend-c")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("backend-a")},
				{Name: types.StringValue("backend-b")},
				{Name: types.StringValue("backend-c")},
			},
		},
		{
			name: "reverse sorted",
			input: []NestedModel{
				{Name: types.StringValue("backend-c")},
				{Name: types.StringValue("backend-b")},
				{Name: types.StringValue("backend-a")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("backend-a")},
				{Name: types.StringValue("backend-b")},
				{Name: types.StringValue("backend-c")},
			},
		},
		{
			name: "unsorted",
			input: []NestedModel{
				{Name: types.StringValue("zebra")},
				{Name: types.StringValue("apple")},
				{Name: types.StringValue("mango")},
				{Name: types.StringValue("banana")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("apple")},
				{Name: types.StringValue("banana")},
				{Name: types.StringValue("mango")},
				{Name: types.StringValue("zebra")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i].Name.ValueString(), result[i].Name.ValueString())
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []NestedModel
		b        []NestedModel
		expected bool
	}{
		{
			name:     "both empty",
			a:        []NestedModel{},
			b:        []NestedModel{},
			expected: true,
		},
		{
			name: "identical single element",
			a: []NestedModel{
				minimalNestedModel(),
			},
			b: []NestedModel{
				minimalNestedModel(),
			},
			expected: true,
		},
		{
			name: "identical multiple elements",
			a: []NestedModel{
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.a.com")},
				{Name: types.StringValue("backend-b"), Address: types.StringValue("api.b.com")},
			},
			b: []NestedModel{
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.a.com")},
				{Name: types.StringValue("backend-b"), Address: types.StringValue("api.b.com")},
			},
			expected: true,
		},
		{
			name: "different order but same content",
			a: []NestedModel{
				{Name: types.StringValue("backend-b"), Address: types.StringValue("api.b.com")},
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.a.com")},
			},
			b: []NestedModel{
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.a.com")},
				{Name: types.StringValue("backend-b"), Address: types.StringValue("api.b.com")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("backend-a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("backend-a")},
				{Name: types.StringValue("backend-b")},
			},
			expected: false,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.a.com")},
			},
			b: []NestedModel{
				{Name: types.StringValue("backend-a"), Address: types.StringValue("api.different.com")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equal(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
