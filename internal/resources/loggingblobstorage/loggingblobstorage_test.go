package loggingblobstorage

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	fwdefaults "github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
)

// Test helpers

func defaultNestedModel() NestedModel {
	return NestedModel{
		commonModel:       defaultCommonModel(),
		Format:            types.StringValue(constants.LoggingBlobStorageDefaultFormat),
		FormatVersion:     types.Int64Value(DefaultFormatVersion),
		Placement:         types.StringNull(),
		ResponseCondition: types.StringValue(DefaultResponseCondition),
	}
}

func defaultCommonModel() commonModel {
	return commonModel{
		Name:             types.StringValue(""),
		Container:        types.StringValue(""),
		Authentication:   NewAuthenticationObject(types.StringValue(""), types.StringValue("")),
		Path:             types.StringValue(DefaultPath),
		Period:           types.Int64Value(DefaultPeriod),
		GzipLevel:        types.Int64Value(DefaultGzipLevel),
		CompressionCodec: types.StringValue(DefaultCompressionCodec),
		MessageType:      types.StringValue(DefaultMessageType),
		TimestampFormat:  types.StringValue(DefaultTimestampFormat),
		FileMaxBytes:     types.Int64Value(DefaultFileMaxBytes),
		PublicKey:        types.StringValue(DefaultPublicKey),
		ProcessingRegion: types.StringValue(DefaultProcessingRegion),
	}
}

func fullNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-blobstorage")
	m.Container = types.StringValue("test-container")
	m.Authentication = NewAuthenticationObject(
		types.StringValue("myaccount"),
		types.StringValue("sv=2020-09-05&sr=b&sig=abc"),
	)
	m.Path = types.StringValue("/logs/")
	m.Period = types.Int64Value(1800)
	m.GzipLevel = types.Int64Value(6)
	m.CompressionCodec = types.StringValue("")
	m.MessageType = types.StringValue("loggly")
	m.TimestampFormat = types.StringValue("%Y")
	m.FileMaxBytes = types.Int64Value(2097152)
	m.PublicKey = types.StringValue("pgp-public-key")
	m.ProcessingRegion = types.StringValue("us")
	m.Format = types.StringValue("%h %l %u")
	m.FormatVersion = types.Int64Value(1)
	m.Placement = types.StringValue("none")
	m.ResponseCondition = types.StringValue("response-condition-1")
	return m
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-blobstorage")
	m.Container = types.StringValue("test-container")
	return m
}

func fullComputeNestedModel() ComputeNestedModel {
	return ComputeNestedModel{commonModel: fullNestedModel().commonModel}
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		bs       *fastly.BlobStorage
		expected NestedModel
	}{
		{
			name:     "nil BlobStorage returns empty model",
			bs:       nil,
			expected: NestedModel{},
		},
		{
			name: "BlobStorage with only required fields uses defaults",
			bs: &fastly.BlobStorage{
				Name:      new("test-blobstorage"),
				Container: new("test-container"),
			},
			expected: minimalNestedModel(),
		},
		{
			name: "BlobStorage with all fields populated",
			bs: &fastly.BlobStorage{
				Name:              new("test-blobstorage"),
				Container:         new("test-container"),
				AccountName:       new("myaccount"),
				SASToken:          new("sv=2020-09-05&sr=b&sig=abc"),
				Path:              new("/logs/"),
				Period:            new(1800),
				GzipLevel:         new(6),
				CompressionCodec:  new(""),
				MessageType:       new("loggly"),
				TimestampFormat:   new("%Y"),
				FileMaxBytes:      new(2097152),
				PublicKey:         new("pgp-public-key"),
				ProcessingRegion:  new("us"),
				Format:            new("%h %l %u"),
				FormatVersion:     new(1),
				Placement:         new("none"),
				ResponseCondition: new("response-condition-1"),
			},
			expected: fullNestedModel(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.bs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenToComputeNestedModel(t *testing.T) {
	bs := &fastly.BlobStorage{
		Name:             new("test-blobstorage"),
		Container:        new("test-container"),
		AccountName:      new("myaccount"),
		SASToken:         new("sv=2020-09-05&sr=b&sig=abc"),
		Path:             new("/logs/"),
		Period:           new(1800),
		GzipLevel:        new(6),
		MessageType:      new("loggly"),
		TimestampFormat:  new("%Y"),
		FileMaxBytes:     new(2097152),
		PublicKey:        new("pgp-public-key"),
		ProcessingRegion: new("us"),
		// VCL-only fields must be ignored by the Compute flatten.
		Format:            new("%h %l %u"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("response-condition-1"),
	}

	result := FlattenToComputeNestedModel(bs)
	assert.Equal(t, fullComputeNestedModel(), result)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		bs       *fastly.BlobStorage
		validate func(t *testing.T, m *Model)
	}{
		{
			name: "nil BlobStorage leaves model untouched",
			bs:   nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "BlobStorage with service metadata builds composite ID",
			bs: &fastly.BlobStorage{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-blobstorage"),
				Container:      new("test-container"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-blobstorage"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-blobstorage"), m.Name)
				assert.Equal(t, types.StringValue("test-container"), m.Container)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.bs, m)
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
		validate  func(t *testing.T, input *fastly.CreateBlobStorageInput)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateBlobStorageInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-blobstorage", *input.Name)
				assert.Equal(t, "test-container", *input.Container)
				assert.Nil(t, input.AccountName)
				assert.Nil(t, input.SASToken)
				assert.Nil(t, input.GzipLevel, "unset gzip_level sentinel should not be sent")
				assert.Nil(t, input.FileMaxBytes)
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateBlobStorageInput) {
				assert.Equal(t, "test-blobstorage", *input.Name)
				assert.Equal(t, "test-container", *input.Container)
				assert.Equal(t, "myaccount", *input.AccountName)
				assert.Equal(t, "sv=2020-09-05&sr=b&sig=abc", *input.SASToken)
				assert.Equal(t, "/logs/", *input.Path)
				assert.Equal(t, 1800, *input.Period)
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 6, *input.GzipLevel)
				assert.Equal(t, "loggly", *input.MessageType)
				assert.Equal(t, "%Y", *input.TimestampFormat)
				assert.NotNil(t, input.FileMaxBytes)
				assert.Equal(t, 2097152, *input.FileMaxBytes)
				assert.Equal(t, "pgp-public-key", *input.PublicKey)
				assert.Equal(t, "us", *input.ProcessingRegion)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, "none", *input.Placement)
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
			validate: func(t *testing.T, input *fastly.CreateBlobStorageInput) {
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 0, *input.GzipLevel)
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
			validate: func(t *testing.T, input *fastly.CreateBlobStorageInput) {
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
	assert.Equal(t, "test-blobstorage", *input.Name)
	assert.Equal(t, "test-container", *input.Container)
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
		validate  func(t *testing.T, input *fastly.UpdateBlobStorageInput)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   6,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.UpdateBlobStorageInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 6, input.ServiceVersion)
				assert.Equal(t, "test-blobstorage", input.Name)
				assert.Equal(t, "test-blobstorage", *input.NewName)
				assert.Equal(t, "test-container", *input.Container)
				assert.Nil(t, input.GzipLevel)
				assert.NotNil(t, input.FileMaxBytes)
				assert.Equal(t, 0, *input.FileMaxBytes)
				assert.Equal(t, fastly.NullValue[string](), input.Placement, "unset placement clears to a JSON null, distinct from an explicit \"none\"")
				assert.NotNil(t, input.CompressionCodec)
				assert.Equal(t, "", *input.CompressionCodec)
				assert.NotNil(t, input.ResponseCondition)
				assert.Equal(t, "", *input.ResponseCondition)
				assert.NotNil(t, input.PublicKey)
				assert.Equal(t, "", *input.PublicKey)
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.UpdateBlobStorageInput) {
				assert.NotNil(t, input.GzipLevel)
				assert.Equal(t, 6, *input.GzipLevel)
				assert.Equal(t, "myaccount", *input.AccountName)
				assert.Equal(t, "sv=2020-09-05&sr=b&sig=abc", *input.SASToken)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, fastly.NewNullable("none"), input.Placement)
				assert.Equal(t, "response-condition-1", *input.ResponseCondition)
				assert.Equal(t, "pgp-public-key", *input.PublicKey)
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
	assert.Equal(t, "test-blobstorage", input.Name)
	assert.Equal(t, "test-blobstorage", *input.NewName)
	assert.Equal(t, 6, *input.GzipLevel)
	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
	assert.Equal(t, "pgp-public-key", *input.PublicKey)
}

func TestClearVCLOnlyCreateFields(t *testing.T) {
	input := &fastly.CreateBlobStorageInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyCreateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyUpdateFields(t *testing.T) {
	input := &fastly.UpdateBlobStorageInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         fastly.NewNullable("none"),
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
		types.StringValue("myaccount"),
		types.StringValue("sv=token"),
	)
	m := commonModel{Authentication: auth}

	assert.Equal(t, types.StringValue("myaccount"), m.AccountName())
	assert.Equal(t, types.StringValue("sv=token"), m.SASToken())
}

func TestAuthenticationAccessorsOnNullObject(t *testing.T) {
	m := commonModel{Authentication: types.ObjectNull(authenticationAttributeTypes)}

	assert.Equal(t, types.StringValue(""), m.AccountName())
	assert.Equal(t, types.StringValue(""), m.SASToken())
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
			name: "different container",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Container = types.StringValue("container-1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Container = types.StringValue("container-2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different authentication",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Authentication = NewAuthenticationObject(types.StringValue("account-1"), types.StringValue(""))
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Authentication = NewAuthenticationObject(types.StringValue("account-2"), types.StringValue(""))
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

	b.Container = types.StringValue("different-container")
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
					m.Container = types.StringValue("different-container")
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

	b[0].Container = types.StringValue("different-container")
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

// TestAuthenticationEnvDefault_DefaultObject covers the schema.Default set on
// the `authentication` attribute itself (schema.go), which is what makes the
// FASTLY_AZURE_* environment variables apply when the whole `authentication`
// block is omitted from config — a plain schema.Default on account_name alone
// is never evaluated in that case, since the framework marks the parent
// object wholly unknown instead of walking into its children.
func TestAuthenticationEnvDefault_DefaultObject(t *testing.T) {
	t.Run("blank when no env vars set", func(t *testing.T) {
		t.Setenv("FASTLY_AZURE_ACCOUNT_NAME", "")
		t.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", "")

		var resp fwdefaults.ObjectResponse
		authenticationEnvDefault{}.DefaultObject(context.Background(), fwdefaults.ObjectRequest{}, &resp)

		assert.Equal(t, NewAuthenticationObject(types.StringValue(""), types.StringValue("")), resp.PlanValue)
	})

	t.Run("account name and sas token from env", func(t *testing.T) {
		t.Setenv("FASTLY_AZURE_ACCOUNT_NAME", "myaccount")
		t.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", "sv=token")

		var resp fwdefaults.ObjectResponse
		authenticationEnvDefault{}.DefaultObject(context.Background(), fwdefaults.ObjectRequest{}, &resp)

		assert.Equal(t, NewAuthenticationObject(
			types.StringValue("myaccount"),
			types.StringValue("sv=token"),
		), resp.PlanValue)
	})
}

// TestResetVCLOnlyToDefaults covers the Compute read-back path. On a Compute
// service the VCL-only fields are never sent, so the API reports its own
// server-side values — notably placement forced to "none" on wasm services.
// Adopting those breaks consistency-after-apply, so they must be reset to
// exactly the values a plan produces.
func TestResetVCLOnlyToDefaults(t *testing.T) {
	m := FlattenToNestedModel(&fastly.BlobStorage{
		Name:              new("test-blobstorage"),
		Container:         new("test-container"),
		ProcessingRegion:  new("none"),
		Format:            new("something-the-server-chose"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("some-condition"),
	})

	ResetVCLOnlyToDefaults(&m)

	assert.Equal(t, constants.LoggingBlobStorageDefaultFormat, m.Format.ValueString())
	assert.Equal(t, int64(DefaultFormatVersion), m.FormatVersion.ValueInt64())
	assert.True(t, m.Placement.IsNull(), "placement must go back to unset, not the API's forced \"none\"")
	assert.Equal(t, DefaultResponseCondition, m.ResponseCondition.ValueString())

	// Non-VCL-only fields must survive untouched.
	assert.Equal(t, "test-blobstorage", m.Name.ValueString())
	assert.Equal(t, "test-container", m.Container.ValueString())
	assert.Equal(t, "none", m.ProcessingRegion.ValueString())
}

// TestResetVCLOnlyToDefaultsMatchesPlannedDefaults ties the reset to the schema
// itself: the values it writes must equal the schema's declared defaults, or
// Create/Update would still disagree with the plan.
func TestResetVCLOnlyToDefaultsMatchesPlannedDefaults(t *testing.T) {
	var m NestedModel
	ResetVCLOnlyToDefaults(&m)

	attrs := CommonAttributes()

	var fResp fwdefaults.StringResponse
	attrs["format"].(schema.StringAttribute).Default.DefaultString(context.Background(), fwdefaults.StringRequest{}, &fResp)
	assert.Equal(t, fResp.PlanValue, m.Format, "format must match its schema default")

	var fvResp fwdefaults.Int64Response
	attrs["format_version"].(schema.Int64Attribute).Default.DefaultInt64(context.Background(), fwdefaults.Int64Request{}, &fvResp)
	assert.Equal(t, fvResp.PlanValue, m.FormatVersion, "format_version must match its schema default")

	var rcResp fwdefaults.StringResponse
	attrs["response_condition"].(schema.StringAttribute).Default.DefaultString(context.Background(), fwdefaults.StringRequest{}, &rcResp)
	assert.Equal(t, rcResp.PlanValue, m.ResponseCondition, "response_condition must match its schema default")

	// placement is Optional-only with no Default, so an absent config value plans
	// as null — the reset has to produce null, not "".
	assert.Nil(t, attrs["placement"].(schema.StringAttribute).Default)
	assert.True(t, m.Placement.IsNull())
}
