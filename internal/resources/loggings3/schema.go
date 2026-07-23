package loggings3

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/defaults"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	fwdefaults "github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultFormatVersion = 2
	// DefaultGzipLevel is a sentinel meaning "gzip_level not configured". A real
	// value is 0-9, so -1 lets the provider distinguish an unset level from an
	// explicit 0 (valid "no compression"). An unset level is never written,
	// because the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level otherwise (e.g. 3 for gzip).
	DefaultGzipLevel                    = -1
	DefaultMessageType                  = "blank"
	DefaultPath                         = ""
	DefaultPeriod                       = 3600
	DefaultTimestampFormat              = "%Y-%m-%dT%H:%M:%S.000"
	DefaultCompressionCodec             = ""
	DefaultPlacement                    = "none"
	DefaultResponseCondition            = ""
	DefaultDomain                       = "s3.amazonaws.com"
	DefaultACL                          = ""
	DefaultRedundancy                   = ""
	DefaultServerSideEncryption         = ""
	DefaultServerSideEncryptionKMSKeyID = ""
	DefaultProcessingRegion             = "none"
	DefaultPublicKey                    = ""
	DefaultFileMaxBytes                 = 0

	// maximumFormatLength is the maximum length the Fastly API accepts for a
	// logging endpoint `format` string. Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	maximumFormatLength = 12288
)

// commonModel holds the S3 logging attributes shared by VCL and Compute
// services. format, format_version, placement, and response_condition only
// affect generated VCL, so they live on NestedModel only — Compute services
// use ComputeNestedModel, which embeds just this common set.
type commonModel struct {
	Name                         types.String `tfsdk:"name"`
	BucketName                   types.String `tfsdk:"bucket_name"`
	Authentication               types.Object `tfsdk:"authentication"`
	Domain                       types.String `tfsdk:"domain"`
	Path                         types.String `tfsdk:"path"`
	Period                       types.Int64  `tfsdk:"period"`
	GzipLevel                    types.Int64  `tfsdk:"gzip_level"`
	CompressionCodec             types.String `tfsdk:"compression_codec"`
	MessageType                  types.String `tfsdk:"message_type"`
	TimestampFormat              types.String `tfsdk:"timestamp_format"`
	ACL                          types.String `tfsdk:"acl"`
	Redundancy                   types.String `tfsdk:"redundancy"`
	ServerSideEncryption         types.String `tfsdk:"server_side_encryption"`
	ServerSideEncryptionKMSKeyID types.String `tfsdk:"server_side_encryption_kms_key_id"`
	FileMaxBytes                 types.Int64  `tfsdk:"file_max_bytes"`
	PublicKey                    types.String `tfsdk:"public_key"`
	ProcessingRegion             types.String `tfsdk:"processing_region"`
}

// NestedModel is the S3 logging model for the standalone
// fastly_service_logging_s3 resource and the VCL nested block
// (service_cdn_auto.logging_s3).
type NestedModel struct {
	commonModel
	Format            types.String `tfsdk:"format"`
	FormatVersion     types.Int64  `tfsdk:"format_version"`
	Placement         types.String `tfsdk:"placement"`
	ResponseCondition types.String `tfsdk:"response_condition"`
}

// ComputeNestedModel is the S3 logging model for the Compute nested block
// (service_compute_auto.logging_s3). It omits format, format_version,
// placement, and response_condition, which only apply to VCL services.
type ComputeNestedModel struct {
	commonModel
}

var authenticationAttributeTypes = map[string]attr.Type{
	"access_key": types.StringType,
	"secret_key": types.StringType,
	"iam_role":   types.StringType,
}

func NewAuthenticationObject(accessKey, secretKey, iamRole types.String) types.Object {
	return types.ObjectValueMust(
		authenticationAttributeTypes,
		map[string]attr.Value{
			"access_key": accessKey,
			"secret_key": secretKey,
			"iam_role":   iamRole,
		},
	)
}

// authenticationEnvDefault populates the authentication object from the
// FASTLY_S3_* environment variables when the practitioner omits the whole
// `authentication` block. The framework only walks into an object
// attribute's per-field Default handlers (like access_key's) once the
// object itself already resolves to a known value; a Computed object
// attribute with no Default of its own is instead marked wholesale
// unknown, and its children's Defaults are never evaluated. Setting this
// Default on the parent gives the object a known value up front so the
// per-field defaults still run for a partially-configured object (e.g.
// only `iam_role` set).
type authenticationEnvDefault struct{}

func (authenticationEnvDefault) Description(_ context.Context) string {
	return "value defaults to the FASTLY_S3_ACCESS_KEY, FASTLY_S3_SECRET_KEY, and FASTLY_S3_IAM_ROLE environment variables"
}

func (d authenticationEnvDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (authenticationEnvDefault) DefaultObject(ctx context.Context, _ fwdefaults.ObjectRequest, resp *fwdefaults.ObjectResponse) {
	resp.PlanValue = NewAuthenticationObject(
		envStringDefault(ctx, "FASTLY_S3_ACCESS_KEY"),
		envStringDefault(ctx, "FASTLY_S3_SECRET_KEY"),
		envStringDefault(ctx, "FASTLY_S3_IAM_ROLE"),
	)
}

func envStringDefault(ctx context.Context, envVar string) types.String {
	var resp fwdefaults.StringResponse
	defaults.EnvString(envVar, "").DefaultString(ctx, fwdefaults.StringRequest{}, &resp)
	return resp.PlanValue
}

func authenticationValue(auth types.Object, name string) types.String {
	if auth.IsNull() || auth.IsUnknown() {
		return types.StringValue("")
	}
	value, ok := auth.Attributes()[name]
	if !ok || value == nil || value.IsNull() || value.IsUnknown() {
		return types.StringValue("")
	}
	stringValue, ok := value.(types.String)
	if !ok {
		return types.StringValue("")
	}
	return stringValue
}

func (n commonModel) AccessKey() types.String {
	return authenticationValue(n.Authentication, "access_key")
}

func (n commonModel) SecretKey() types.String {
	return authenticationValue(n.Authentication, "secret_key")
}

func (n commonModel) IAMRole() types.String {
	return authenticationValue(n.Authentication, "iam_role")
}

func (n commonModel) equal(other commonModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.BucketName) == service.StringValue(other.BucketName) &&
		service.StringValue(n.AccessKey()) == service.StringValue(other.AccessKey()) &&
		service.StringValue(n.SecretKey()) == service.StringValue(other.SecretKey()) &&
		service.StringValue(n.IAMRole()) == service.StringValue(other.IAMRole()) &&
		service.StringValue(n.Domain) == service.StringValue(other.Domain) &&
		service.StringValue(n.Path) == service.StringValue(other.Path) &&
		service.Int64Value(n.Period) == service.Int64Value(other.Period) &&
		service.Int64Value(n.GzipLevel) == service.Int64Value(other.GzipLevel) &&
		service.StringValue(n.CompressionCodec) == service.StringValue(other.CompressionCodec) &&
		service.StringValue(n.MessageType) == service.StringValue(other.MessageType) &&
		service.StringValue(n.TimestampFormat) == service.StringValue(other.TimestampFormat) &&
		service.StringValue(n.ACL) == service.StringValue(other.ACL) &&
		service.StringValue(n.Redundancy) == service.StringValue(other.Redundancy) &&
		service.StringValue(n.ServerSideEncryption) == service.StringValue(other.ServerSideEncryption) &&
		service.StringValue(n.ServerSideEncryptionKMSKeyID) == service.StringValue(other.ServerSideEncryptionKMSKeyID) &&
		service.Int64Value(n.FileMaxBytes) == service.Int64Value(other.FileMaxBytes) &&
		service.StringValue(n.PublicKey) == service.StringValue(other.PublicKey) &&
		service.StringValue(n.ProcessingRegion) == service.StringValue(other.ProcessingRegion)
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return n.commonModel.equal(other.commonModel) &&
		service.StringValue(n.Format) == service.StringValue(other.Format) &&
		service.Int64Value(n.FormatVersion) == service.Int64Value(other.FormatVersion) &&
		service.StringValue(n.Placement) == service.StringValue(other.Placement) &&
		service.StringValue(n.ResponseCondition) == service.StringValue(other.ResponseCondition)
}

func (c ComputeNestedModel) ModelsEqual(other ComputeNestedModel) bool {
	return c.commonModel.equal(other.commonModel)
}

// CommonAttributes returns the full S3 logging attribute set — the shared
// attributes plus the VCL-only ones (format, format_version, placement,
// response_condition). Used by the standalone fastly_service_logging_s3
// resource (which can attach to either service type) and the VCL nested
// block (NestedBlockSchema). Compute services use ComputeAttributes instead.
func CommonAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	maps.Copy(attrs, vclOnlyAttributes())
	return attrs
}

// ComputeAttributes returns the S3 logging attribute set for Compute
// services, omitting the VCL-only attributes exposed by CommonAttributes.
func ComputeAttributes() map[string]schema.Attribute {
	return sharedAttributes()
}

// sharedAttributes returns the S3 logging attributes common to both VCL and
// Compute services.
func sharedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		// Required
		"bucket_name": schema.StringAttribute{
			Required:    true,
			Description: "The bucket name for S3 account.",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name for the real-time logging configuration. Must be unique within the service.",
		},
		// Optional
		"acl": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultACL),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(fastly.S3AccessControlListPrivate),
					string(fastly.S3AccessControlListPublicRead),
					string(fastly.S3AccessControlListPublicReadWrite),
					string(fastly.S3AccessControlListAWSExecRead),
					string(fastly.S3AccessControlListAuthenticatedRead),
					string(fastly.S3AccessControlListBucketOwnerRead),
					string(fastly.S3AccessControlListBucketOwnerFullControl),
				),
			},
			Description: "The access control list (ACL) specific request header. See the AWS documentation for [Access Control List (ACL) Specific Request Headers](https://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadInitiate.html#initiate-mpu-acl-specific-request-headers) for more information.",
		},
		"authentication": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Default:     authenticationEnvDefault{},
			Description: "AWS authentication credentials for S3 access. Provide either `access_key` and `secret_key`, or `iam_role`. When this block is omitted entirely, defaults to the `FASTLY_S3_ACCESS_KEY`, `FASTLY_S3_SECRET_KEY`, and `FASTLY_S3_IAM_ROLE` environment variables.",
			Attributes: map[string]schema.Attribute{
				"access_key": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     defaults.EnvString("FASTLY_S3_ACCESS_KEY", ""),
					Description: "The access key for your S3 account. Not required if `iam_role` is provided. Can be set via the `FASTLY_S3_ACCESS_KEY` environment variable.",
				},
				"iam_role": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     defaults.EnvString("FASTLY_S3_IAM_ROLE", ""),
					Description: "The Amazon Resource Name (ARN) for the IAM role granting Fastly access to S3. Not required if `access_key` and `secret_key` are provided. Can be set via the `FASTLY_S3_IAM_ROLE` environment variable.",
				},
				"secret_key": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     defaults.EnvString("FASTLY_S3_SECRET_KEY", ""),
					Description: "The secret key for your S3 account. Not required if `iam_role` is provided. Can be set via the `FASTLY_S3_SECRET_KEY` environment variable.",
				},
			},
		},
		"compression_codec": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultCompressionCodec),
			Validators: []validator.String{
				stringvalidator.OneOf("zstd", "snappy", "gzip"),
			},
			Description: "The codec used for compressing your logs. Valid values are `zstd`, `snappy`, and `gzip`. If the codec is `gzip`, `gzip_level` defaults to `3`; to use a different level, leave `compression_codec` unset and set `gzip_level` instead. Conflicts with `gzip_level`: setting both in the same request will result in an error.",
		},
		"domain": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultDomain),
			Description: "The Domain of the Amazon S3 endpoint.",
		},
		"file_max_bytes": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultFileMaxBytes),
			Description: "The maximum number of bytes for each uploaded file. A value of 0 can be used to indicate there is no limit on the size of uploaded files, otherwise the minimum value is 1048576 bytes (1 MiB.).",
		},
		"gzip_level": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultGzipLevel),
			// compression_codec and gzip_level are mutually exclusive; the API
			// rejects a request that sets both. Validation runs against config,
			// where an unset gzip_level is null rather than the -1 default, so
			// this correctly fires only when both are set.
			Validators: []validator.Int64{
				gzipLevelCodecConflict{},
			},
			Description: "The level of gzip encoding when sending logs. Valid values are `0` (no compression) through `9`. To compress at a specific gzip level, leave `compression_codec` unset and set this. Conflicts with `compression_codec`: setting both in the same request will result in an error.",
		},
		"message_type": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultMessageType),
			Validators: []validator.String{
				stringvalidator.OneOf("classic", "loggly", "logplex", "blank"),
			},
			Description: "How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `blank`.",
		},
		"path": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultPath),
			Description: "Path to store the files. Must end with a trailing slash. If this field is left empty, the files will be saved in the bucket's root path.",
		},
		"period": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultPeriod),
			Description: "How frequently log files are finalized so they can be available for reading in seconds. Default `3600`.",
		},
		"processing_region": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultProcessingRegion),
			Validators: []validator.String{
				stringvalidator.OneOf("none", "us", "eu"),
			},
			Description: "Region where logs will be processed before streaming to the destination. Valid values are `none`, `us` and `eu`.",
		},
		"public_key": schema.StringAttribute{
			Optional:  true,
			Computed:  true,
			Sensitive: true,
			Default:   stringdefault.StaticString(DefaultPublicKey),
			Validators: []validator.String{
				notTrimmed{},
			},
			Description: "PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
		},
		"redundancy": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultRedundancy),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(fastly.S3RedundancyStandard),
					string(fastly.S3RedundancyIntelligentTiering),
					string(fastly.S3RedundancyStandardIA),
					string(fastly.S3RedundancyOneZoneIA),
					string(fastly.S3RedundancyGlacierFlexibleRetrieval),
					string(fastly.S3RedundancyGlacierInstantRetrieval),
					string(fastly.S3RedundancyGlacierDeepArchive),
					string(fastly.S3RedundancyReduced),
				),
			},
			Description: "The S3 redundancy level. Valid values are `standard`, `intelligent_tiering`, `standard_ia`, `onezone_ia`, `glacier_ir`, `glacier`, `deep_archive`, and `reduced_redundancy`.",
		},
		"server_side_encryption": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultServerSideEncryption),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(fastly.S3ServerSideEncryptionAES),
					string(fastly.S3ServerSideEncryptionKMS),
				),
			},
			Description: "Server-side encryption method. Valid values are `AES256` and `aws:kms`.",
		},
		"server_side_encryption_kms_key_id": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultServerSideEncryptionKMSKeyID),
			Description: "KMS key ID to use for `server_side_encryption`. Required when `server_side_encryption` is `aws:kms`.",
		},
		"timestamp_format": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultTimestampFormat),
			Description: "strftime-specified timestamp format for log filename.",
		},
	}
}

// vclOnlyAttributes returns the S3 logging attributes that only affect
// generated VCL and have no meaning for Compute services.
func vclOnlyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(constants.LoggingS3DefaultFormat),
			Validators: []validator.String{
				stringvalidator.LengthAtMost(maximumFormatLength),
			},
			Description: "A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/).",
		},
		"format_version": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultFormatVersion),
			Validators: []validator.Int64{
				int64validator.Between(1, 2),
			},
			Description: "The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in vcl_log if format_version is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.",
		},
		"placement": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultPlacement),
			Validators: []validator.String{
				stringvalidator.OneOf("none"),
			},
			Description: "Where in the generated VCL the logging call should be placed. If not set, endpoints with format_version of 2 are placed in vcl_log and those with format_version of 1 are placed in vcl_deliver. Valid value is `none`.",
		},
		"response_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultResponseCondition),
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
		},
	}
	maps.Copy(attrs, CommonAttributes())
	// For the standalone resource, service_id + name locate the endpoint in the
	// API, so a change to either cannot be an in-place update. version is not
	// replacement-forcing: the explicit clone workflow copies the endpoint into
	// the new version, so an in-place update there succeeds. Applied to name
	// here (not in CommonAttributes) so the nested block, where name is only a
	// list key, is unaffected.
	nameAttr := attrs["name"].(schema.StringAttribute)
	nameAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["name"] = nameAttr
	return attrs
}

// NestedBlockSchema returns the S3 logging nested block schema for VCL
// services (service_cdn_auto.logging_s3), including the VCL-only attributes.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "S3 logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ComputeNestedBlockSchema returns the S3 logging nested block schema for
// Compute services (service_compute_auto.logging_s3), omitting the VCL-only
// attributes.
func ComputeNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "S3 logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: ComputeAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.S3, error) {
	return client.ListS3s(ctx, &fastly.ListS3sInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.S3) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteS3(ctx, &fastly.DeleteS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.S3, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateS3(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.S3) bool {
	remoteModel := FlattenToNestedModel(remote)
	preserveGzipSentinel(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.S3, error) {
	input := BuildUpdateInput(serviceID, version, desired)
	return client.UpdateS3(ctx, input)
}

func (o ops) ToModel(api *fastly.S3) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.S3]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return reconciler.ReadForVersion(ctx, client, serviceID, version)
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	result := reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
	// order carries the configured/prior models, so it holds the gzip_level
	// sentinel for endpoints the user left unset; preserve it on the read-back.
	preserveGzipSentinelList(result, order)
	return result
}

type computeOps struct{}

func (o computeOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.S3, error) {
	return client.ListS3s(ctx, &fastly.ListS3sInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o computeOps) GetName(api *fastly.S3) string {
	return fastly.ToValue(api.Name)
}

func (o computeOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteS3(ctx, &fastly.DeleteS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o computeOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.S3, error) {
	input := BuildComputeCreateInput(serviceID, version, desired)
	return client.CreateS3(ctx, input)
}

func (o computeOps) Equal(desired ComputeNestedModel, remote *fastly.S3) bool {
	remoteModel := FlattenToComputeNestedModel(remote)
	preserveGzipSentinelCompute(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o computeOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.S3, error) {
	input := BuildComputeUpdateInput(serviceID, version, desired)
	return client.UpdateS3(ctx, input)
}

func (o computeOps) ToModel(api *fastly.S3) ComputeNestedModel {
	return FlattenToComputeNestedModel(api)
}

var computeReconciler = &reconcile.Resource[ComputeNestedModel, fastly.S3]{
	Ops: computeOps{},
	GetName: func(m ComputeNestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

func ComputeReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]ComputeNestedModel, error) {
	return computeReconciler.ReadForVersion(ctx, client, serviceID, version)
}

func ComputeReconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []ComputeNestedModel) error {
	return computeReconciler.Run(ctx, client, serviceID, version, desired)
}

func ComputeEqual(a, b []ComputeNestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m ComputeNestedModel) string { return service.StringValue(m.Name) }, ComputeNestedModel.ModelsEqual, true)
}

func ComputeMatchOrder(items, order []ComputeNestedModel) []ComputeNestedModel {
	result := reconcile.MatchOrder(items, order, func(m ComputeNestedModel) string { return service.StringValue(m.Name) })
	// order carries the configured/prior models, so it holds the gzip_level
	// sentinel for endpoints the user left unset; preserve it on the read-back.
	preserveGzipSentinelListCompute(result, order)
	return result
}
