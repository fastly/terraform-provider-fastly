package loggings3

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
)

// Test helpers

func defaultNestedModel() NestedModel {
	return NestedModel{
		commonModel:       defaultCommonModel(),
		Format:            types.StringValue(constants.LoggingS3DefaultFormat),
		FormatVersion:     types.Int64Value(DefaultFormatVersion),
		Placement:         types.StringValue(DefaultPlacement),
		ResponseCondition: types.StringValue(DefaultResponseCondition),
	}
}

func defaultCommonModel() commonModel {
	return commonModel{
		Name:                         types.StringValue(""),
		BucketName:                   types.StringValue(""),
		Authentication:               NewAuthenticationObject(types.StringValue(""), types.StringValue(""), types.StringValue("")),
		Domain:                       types.StringValue(DefaultDomain),
		Path:                         types.StringValue(DefaultPath),
		Period:                       types.Int64Value(DefaultPeriod),
		GzipLevel:                    types.Int64Value(DefaultGzipLevel),
		CompressionCodec:             types.StringValue(DefaultCompressionCodec),
		MessageType:                  types.StringValue(DefaultMessageType),
		TimestampFormat:              types.StringValue(DefaultTimestampFormat),
		ACL:                          types.StringValue(DefaultACL),
		Redundancy:                   types.StringValue(DefaultRedundancy),
		ServerSideEncryption:         types.StringValue(DefaultServerSideEncryption),
		ServerSideEncryptionKMSKeyID: types.StringValue(DefaultServerSideEncryptionKMSKeyID),
		FileMaxBytes:                 types.Int64Value(DefaultFileMaxBytes),
		PublicKey:                    types.StringValue(DefaultPublicKey),
		ProcessingRegion:             types.StringValue(DefaultProcessingRegion),
	}
}

func fullNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-s3")
	m.BucketName = types.StringValue("test-bucket")
	m.Authentication = NewAuthenticationObject(
		types.StringValue("access-key"),
		types.StringValue("secret-key"),
		types.StringValue("arn:aws:iam::123456789012:role/test"),
	)
	m.Domain = types.StringValue("s3-us-west-2.amazonaws.com")
	m.Path = types.StringValue("/logs/")
	m.Period = types.Int64Value(1800)
	m.GzipLevel = types.Int64Value(6)
	m.CompressionCodec = types.StringValue("")
	m.MessageType = types.StringValue("classic")
	m.TimestampFormat = types.StringValue("%Y")
	m.ACL = types.StringValue("private")
	m.Redundancy = types.StringValue("standard")
	m.ServerSideEncryption = types.StringValue("aws:kms")
	m.ServerSideEncryptionKMSKeyID = types.StringValue("kms-key-1")
	m.FileMaxBytes = types.Int64Value(2097152)
	m.PublicKey = types.StringValue("pgp-public-key")
	m.ProcessingRegion = types.StringValue("us")
	m.Format = types.StringValue("%h %l %u")
	m.FormatVersion = types.Int64Value(1)
	m.Placement = types.StringValue("waf_debug")
	m.ResponseCondition = types.StringValue("response-condition-1")
	return m
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-s3")
	m.BucketName = types.StringValue("test-bucket")
	return m
}

func fullComputeNestedModel() ComputeNestedModel {
	return ComputeNestedModel{commonModel: fullNestedModel().commonModel}
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		s3       *fastly.S3
		expected NestedModel
	}{
		{
			name:     "nil S3 returns empty model",
			s3:       nil,
			expected: NestedModel{},
		},
		{
			name: "S3 with only required fields uses defaults",
			s3: &fastly.S3{
				Name:       new("test-s3"),
				BucketName: new("test-bucket"),
			},
			expected: minimalNestedModel(),
		},
		{
			name: "S3 with all fields populated",
			s3: &fastly.S3{
				Name:                         new("test-s3"),
				BucketName:                   new("test-bucket"),
				AccessKey:                    new("access-key"),
				SecretKey:                    new("secret-key"),
				IAMRole:                      new("arn:aws:iam::123456789012:role/test"),
				Domain:                       new("s3-us-west-2.amazonaws.com"),
				Path:                         new("/logs/"),
				Period:                       new(1800),
				GzipLevel:                    new(6),
				CompressionCodec:             new(""),
				MessageType:                  new("classic"),
				TimestampFormat:              new("%Y"),
				ACL:                          new(fastly.S3AccessControlList("private")),
				Redundancy:                   new(fastly.S3Redundancy("standard")),
				ServerSideEncryption:         new(fastly.S3ServerSideEncryption("aws:kms")),
				ServerSideEncryptionKMSKeyID: new("kms-key-1"),
				FileMaxBytes:                 new(2097152),
				PublicKey:                    new("pgp-public-key"),
				ProcessingRegion:             new("us"),
				Format:                       new("%h %l %u"),
				FormatVersion:                new(1),
				Placement:                    new("waf_debug"),
				ResponseCondition:            new("response-condition-1"),
			},
			expected: fullNestedModel(),
		},
		{
			name: "nil ACL, redundancy, and server side encryption use defaults",
			s3: &fastly.S3{
				Name:                 new("test-s3"),
				BucketName:           new("test-bucket"),
				ACL:                  nil,
				Redundancy:           nil,
				ServerSideEncryption: nil,
			},
			expected: minimalNestedModel(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.s3)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenToComputeNestedModel(t *testing.T) {
	s3 := &fastly.S3{
		Name:                         new("test-s3"),
		BucketName:                   new("test-bucket"),
		AccessKey:                    new("access-key"),
		SecretKey:                    new("secret-key"),
		IAMRole:                      new("arn:aws:iam::123456789012:role/test"),
		Domain:                       new("s3-us-west-2.amazonaws.com"),
		Path:                         new("/logs/"),
		Period:                       new(1800),
		GzipLevel:                    new(6),
		MessageType:                  new("classic"),
		TimestampFormat:              new("%Y"),
		ACL:                          new(fastly.S3AccessControlList("private")),
		Redundancy:                   new(fastly.S3Redundancy("standard")),
		ServerSideEncryption:         new(fastly.S3ServerSideEncryption("aws:kms")),
		ServerSideEncryptionKMSKeyID: new("kms-key-1"),
		FileMaxBytes:                 new(2097152),
		PublicKey:                    new("pgp-public-key"),
		ProcessingRegion:             new("us"),
		// VCL-only fields must be ignored by the Compute flatten.
		Format:            new("%h %l %u"),
		FormatVersion:     new(1),
		Placement:         new("waf_debug"),
		ResponseCondition: new("response-condition-1"),
	}

	result := FlattenToComputeNestedModel(s3)
	assert.Equal(t, fullComputeNestedModel(), result)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		s3       *fastly.S3
		validate func(t *testing.T, m *Model)
	}{
		{
			name: "nil S3 leaves model untouched",
			s3:   nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "S3 with service metadata builds composite ID",
			s3: &fastly.S3{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-s3"),
				BucketName:     new("test-bucket"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-s3"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-s3"), m.Name)
				assert.Equal(t, types.StringValue("test-bucket"), m.BucketName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.s3, m)
			tt.validate(t, m)
		})
	}
}

func TestPreserveGzipSentinel(t *testing.T) {
	tests := []struct {
		name     string
		remote   NestedModel
		desired  NestedModel
		expected types.Int64
	}{
		{
			name: "desired unset restores sentinel over API auto-managed value",
			remote: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(3)
				return m
			}(),
			desired:  minimalNestedModel(),
			expected: types.Int64Value(DefaultGzipLevel),
		},
		{
			name: "desired set keeps the API value",
			remote: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(6)
				return m
			}(),
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(6)
				return m
			}(),
			expected: types.Int64Value(6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.remote
			preserveGzipSentinel(&m, tt.desired)
			assert.Equal(t, tt.expected, m.GzipLevel)
		})
	}
}

func TestInferGzipSentinelOnImport(t *testing.T) {
	tests := []struct {
		name     string
		m        commonModel
		expected types.Int64
	}{
		{
			name: "no codec, gzip_level 0 treated as unconfigured",
			m: commonModel{
				CompressionCodec: types.StringValue(""),
				GzipLevel:        types.Int64Value(0),
			},
			expected: types.Int64Value(DefaultGzipLevel),
		},
		{
			name: "codec set, gzip_level 0 is left alone",
			m: commonModel{
				CompressionCodec: types.StringValue("zstd"),
				GzipLevel:        types.Int64Value(0),
			},
			expected: types.Int64Value(0),
		},
		{
			name: "non-zero gzip_level is left alone",
			m: commonModel{
				CompressionCodec: types.StringValue(""),
				GzipLevel:        types.Int64Value(5),
			},
			expected: types.Int64Value(5),
		},
		{
			name: "already-unset sentinel is left alone",
			m: commonModel{
				CompressionCodec: types.StringValue(""),
				GzipLevel:        types.Int64Value(DefaultGzipLevel),
			},
			expected: types.Int64Value(DefaultGzipLevel),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			inferGzipSentinelOnImport(&m)
			assert.Equal(t, tt.expected, m.GzipLevel)
		})
	}
}

func TestPreserveGzipSentinelList(t *testing.T) {
	read := []NestedModel{
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("a")
			m.GzipLevel = types.Int64Value(3)
			return m
		}(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("b")
			m.GzipLevel = types.Int64Value(6)
			return m
		}(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("c")
			m.GzipLevel = types.Int64Value(0)
			return m
		}(),
	}
	desired := []NestedModel{
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("a")
			return m
		}(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("b")
			m.GzipLevel = types.Int64Value(6)
			return m
		}(),
		// "c" has no entry in desired, simulating a freshly imported or
		// undiscovered-in-config endpoint.
	}

	preserveGzipSentinelList(read, desired)

	assert.Equal(t, types.Int64Value(DefaultGzipLevel), read[2].GzipLevel, "unmatched entry falls back to the import heuristic")

	assert.Equal(t, types.Int64Value(DefaultGzipLevel), read[0].GzipLevel, "unmatched-by-desired sentinel should be restored")
	assert.Equal(t, types.Int64Value(6), read[1].GzipLevel, "explicitly configured value should be preserved")
}

// Tests for expand.go

func TestBuildCreateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.CreateS3Input)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateS3Input) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-s3", *input.Name)
				assert.Equal(t, "test-bucket", *input.BucketName)
				assert.Nil(t, input.AccessKey)
				assert.Nil(t, input.SecretKey)
				assert.Nil(t, input.IAMRole)
				assert.Nil(t, input.GzipLevel, "unset gzip_level sentinel should not be sent")
				assert.Nil(t, input.ACL)
				assert.Nil(t, input.Redundancy)
				assert.Nil(t, input.ServerSideEncryption)
				assert.Nil(t, input.FileMaxBytes)
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateS3Input) {
				assert.Equal(t, "test-s3", *input.Name)
				assert.Equal(t, "test-bucket", *input.BucketName)
				assert.Equal(t, "access-key", *input.AccessKey)
				assert.Equal(t, "secret-key", *input.SecretKey)
				assert.Equal(t, "arn:aws:iam::123456789012:role/test", *input.IAMRole)
				assert.Equal(t, "s3-us-west-2.amazonaws.com", *input.Domain)
				assert.Equal(t, "/logs/", *input.Path)
				assert.Equal(t, 1800, *input.Period)
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 6, *input.GzipLevel)
				assert.Equal(t, "classic", *input.MessageType)
				assert.Equal(t, "%Y", *input.TimestampFormat)
				assert.NotNil(t, input.ACL)
				assert.Equal(t, fastly.S3AccessControlList("private"), *input.ACL)
				assert.NotNil(t, input.Redundancy)
				assert.Equal(t, fastly.S3Redundancy("standard"), *input.Redundancy)
				assert.NotNil(t, input.ServerSideEncryption)
				assert.Equal(t, fastly.S3ServerSideEncryption("aws:kms"), *input.ServerSideEncryption)
				assert.Equal(t, "kms-key-1", *input.ServerSideEncryptionKMSKeyID)
				assert.NotNil(t, input.FileMaxBytes)
				assert.Equal(t, 2097152, *input.FileMaxBytes)
				assert.Equal(t, "pgp-public-key", *input.PublicKey)
				assert.Equal(t, "us", *input.ProcessingRegion)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, "waf_debug", *input.Placement)
				assert.Equal(t, "response-condition-1", *input.ResponseCondition)
			},
		},
		{
			name:      "gzip_level 0 is explicit and sent",
			serviceID: "service-789",
			version:   1,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(0)
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateS3Input) {
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 0, *input.GzipLevel)
			},
		},
		{
			name:      "empty ACL, redundancy, encryption stay unset",
			serviceID: "service-abc",
			version:   2,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateS3Input) {
				assert.Nil(t, input.ACL)
				assert.Nil(t, input.Redundancy)
				assert.Nil(t, input.ServerSideEncryption)
			},
		},
		{
			name:      "file_max_bytes 0 stays unset",
			serviceID: "service-def",
			version:   3,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.FileMaxBytes = types.Int64Value(0)
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateS3Input) {
				assert.Nil(t, input.FileMaxBytes)
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

func TestBuildComputeCreateInput(t *testing.T) {
	input := BuildComputeCreateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-s3", *input.Name)
	assert.Equal(t, "test-bucket", *input.BucketName)
	assert.Equal(t, 6, *input.GzipLevel)
	assert.Nil(t, input.Format, "VCL-only fields must never be set for Compute")
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestBuildUpdateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.UpdateS3Input)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   6,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.UpdateS3Input) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 6, input.ServiceVersion)
				assert.Equal(t, "test-s3", input.Name)
				assert.Equal(t, "test-s3", *input.NewName)
				assert.Equal(t, "test-bucket", *input.BucketName)
				assert.Nil(t, input.GzipLevel)
				assert.NotNil(t, input.ACL)
				assert.Equal(t, fastly.S3AccessControlList(""), *input.ACL)
				assert.NotNil(t, input.Redundancy)
				assert.Equal(t, fastly.S3Redundancy(""), *input.Redundancy)
				assert.Nil(t, input.ServerSideEncryption)
				assert.NotNil(t, input.FileMaxBytes)
				assert.Equal(t, 0, *input.FileMaxBytes)
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.UpdateS3Input) {
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 6, *input.GzipLevel)
				assert.Equal(t, fastly.S3AccessControlList("private"), *input.ACL)
				assert.Equal(t, fastly.S3Redundancy("standard"), *input.Redundancy)
				assert.NotNil(t, input.ServerSideEncryption)
				assert.Equal(t, fastly.S3ServerSideEncryption("aws:kms"), *input.ServerSideEncryption)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, "waf_debug", *input.Placement)
				assert.Equal(t, "response-condition-1", *input.ResponseCondition)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildUpdateInput(tt.serviceID, tt.version, tt.model)
			tt.validate(t, input)
		})
	}
}

func TestBuildComputeUpdateInput(t *testing.T) {
	input := BuildComputeUpdateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-s3", input.Name)
	assert.Equal(t, "test-s3", *input.NewName)
	assert.Equal(t, 6, *input.GzipLevel)
	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyCreateFields(t *testing.T) {
	input := &fastly.CreateS3Input{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("waf_debug"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyCreateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyUpdateFields(t *testing.T) {
	input := &fastly.UpdateS3Input{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("waf_debug"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyUpdateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

// Tests for schema.go

func TestAuthenticationAccessors(t *testing.T) {
	auth := NewAuthenticationObject(
		types.StringValue("access-key"),
		types.StringValue("secret-key"),
		types.StringValue("iam-role"),
	)
	m := commonModel{Authentication: auth}

	assert.Equal(t, types.StringValue("access-key"), m.AccessKey())
	assert.Equal(t, types.StringValue("secret-key"), m.SecretKey())
	assert.Equal(t, types.StringValue("iam-role"), m.IAMRole())
}

func TestAuthenticationAccessorsOnNullObject(t *testing.T) {
	m := commonModel{Authentication: types.ObjectNull(authenticationAttributeTypes)}

	assert.Equal(t, types.StringValue(""), m.AccessKey())
	assert.Equal(t, types.StringValue(""), m.SecretKey())
	assert.Equal(t, types.StringValue(""), m.IAMRole())
}

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
			name: "different bucket name",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.BucketName = types.StringValue("bucket-1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.BucketName = types.StringValue("bucket-2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different authentication",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Authentication = NewAuthenticationObject(types.StringValue("key-1"), types.StringValue(""), types.StringValue(""))
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Authentication = NewAuthenticationObject(types.StringValue("key-2"), types.StringValue(""), types.StringValue(""))
				return m
			}(),
			expected: false,
		},
		{
			name: "different format only affects NestedModel equality",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Format = types.StringValue("format-a")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Format = types.StringValue("format-b")
				return m
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.a.ModelsEqual(tt.b))
		})
	}
}

func TestComputeModelsEqual(t *testing.T) {
	a := fullComputeNestedModel()
	b := fullComputeNestedModel()
	assert.True(t, a.ModelsEqual(b))

	b.BucketName = types.StringValue("different-bucket")
	assert.False(t, a.ModelsEqual(b))
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
			name: "different order but same content matches by name",
			a: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }(),
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
			},
			b: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }(),
			},
			expected: true,
		},
		{
			name: "different content",
			a: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
			},
			b: []NestedModel{
				func() NestedModel {
					m := minimalNestedModel()
					m.Name = types.StringValue("a")
					m.BucketName = types.StringValue("different-bucket")
					return m
				}(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Equal(tt.a, tt.b))
		})
	}
}

func TestComputeEqual(t *testing.T) {
	a := []ComputeNestedModel{fullComputeNestedModel()}
	b := []ComputeNestedModel{fullComputeNestedModel()}
	assert.True(t, ComputeEqual(a, b))

	b[0].BucketName = types.StringValue("different-bucket")
	assert.False(t, ComputeEqual(a, b))
}

func TestMatchOrder(t *testing.T) {
	itemA := func() NestedModel {
		m := minimalNestedModel()
		m.Name = types.StringValue("a")
		m.GzipLevel = types.Int64Value(3)
		return m
	}()
	itemB := func() NestedModel {
		m := minimalNestedModel()
		m.Name = types.StringValue("b")
		m.GzipLevel = types.Int64Value(6)
		return m
	}()
	items := []NestedModel{itemB, itemA}

	orderA := minimalNestedModel()
	orderA.Name = types.StringValue("a")
	orderB := minimalNestedModel()
	orderB.Name = types.StringValue("b")
	orderB.GzipLevel = types.Int64Value(6)
	order := []NestedModel{orderA, orderB}

	result := MatchOrder(items, order)

	assert.Len(t, result, 2)
	assert.Equal(t, "a", result[0].Name.ValueString())
	assert.Equal(t, "b", result[1].Name.ValueString())
	assert.Equal(t, types.Int64Value(DefaultGzipLevel), result[0].GzipLevel, "gzip sentinel restored for unset order entry")
	assert.Equal(t, types.Int64Value(6), result[1].GzipLevel, "gzip value preserved for explicitly configured order entry")
}

func TestComputeMatchOrder(t *testing.T) {
	itemA := func() ComputeNestedModel {
		m := fullComputeNestedModel()
		m.Name = types.StringValue("a")
		m.GzipLevel = types.Int64Value(3)
		return m
	}()
	orderA := ComputeNestedModel{commonModel: minimalNestedModel().commonModel}
	orderA.Name = types.StringValue("a")

	result := ComputeMatchOrder([]ComputeNestedModel{itemA}, []ComputeNestedModel{orderA})

	assert.Len(t, result, 1)
	assert.Equal(t, types.Int64Value(DefaultGzipLevel), result[0].GzipLevel)
}
