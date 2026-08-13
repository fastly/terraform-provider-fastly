package loggingblobstorage

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/defaults"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
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
	DefaultGzipLevel = -1
	// DefaultMessageType matches the Fastly API's own default for this endpoint
	// specifically, which differs from the generic logging default of "blank".
	DefaultMessageType       = "classic"
	DefaultPath              = ""
	DefaultPeriod            = 3600
	DefaultTimestampFormat   = "%Y-%m-%dT%H:%M:%S.000"
	DefaultCompressionCodec  = ""
	DefaultResponseCondition = ""
	DefaultProcessingRegion  = "none"
	DefaultPublicKey         = ""
	DefaultFileMaxBytes      = 0

	// minimumFileMaxBytes is the smallest non-zero file_max_bytes the Fastly API
	// accepts; 0 itself means "no limit". Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	minimumFileMaxBytes = 1048576

	// maximumFormatLength is the maximum length the Fastly API accepts for a
	// logging endpoint `format` string. Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	maximumFormatLength = 12288
)

// commonModel holds the Blob Storage logging attributes shared by VCL and
// Compute services. format, format_version, placement, and response_condition
// only affect generated VCL, so they live on NestedModel only — Compute
// services use ComputeNestedModel, which embeds just this common set.
type commonModel struct {
	Name             types.String `tfsdk:"name"`
	Container        types.String `tfsdk:"container"`
	Authentication   types.Object `tfsdk:"authentication"`
	Path             types.String `tfsdk:"path"`
	Period           types.Int64  `tfsdk:"period"`
	GzipLevel        types.Int64  `tfsdk:"gzip_level"`
	CompressionCodec types.String `tfsdk:"compression_codec"`
	MessageType      types.String `tfsdk:"message_type"`
	TimestampFormat  types.String `tfsdk:"timestamp_format"`
	FileMaxBytes     types.Int64  `tfsdk:"file_max_bytes"`
	PublicKey        types.String `tfsdk:"public_key"`
	ProcessingRegion types.String `tfsdk:"processing_region"`
}

// NestedModel is the Blob Storage logging model for the standalone
// fastly_service_logging_blobstorage resource and the VCL nested block
// (service_cdn_auto.logging_blobstorage).
type NestedModel struct {
	commonModel
	Format            types.String `tfsdk:"format"`
	FormatVersion     types.Int64  `tfsdk:"format_version"`
	Placement         types.String `tfsdk:"placement"`
	ResponseCondition types.String `tfsdk:"response_condition"`
}

// ComputeNestedModel is the Blob Storage logging model for the Compute nested
// block (service_compute_auto.logging_blobstorage). It omits format,
// format_version, placement, and response_condition, which only apply to VCL
// services.
type ComputeNestedModel struct {
	commonModel
}

var authenticationAttributeTypes = map[string]attr.Type{
	"account_name": types.StringType,
	"sas_token":    types.StringType,
}

func NewAuthenticationObject(accountName, sasToken types.String) types.Object {
	return types.ObjectValueMust(
		authenticationAttributeTypes,
		map[string]attr.Value{
			"account_name": accountName,
			"sas_token":    sasToken,
		},
	)
}

// authenticationEnvDefault populates the authentication object from the
// FASTLY_AZURE_ACCOUNT_NAME and FASTLY_AZURE_SHARED_ACCESS_SIGNATURE
// environment variables when the practitioner omits the whole
// `authentication` block. The framework only walks into an object
// attribute's per-field Default handlers (like sas_token's) once the object
// itself already resolves to a known value; a Computed object attribute with
// no Default of its own is instead marked wholesale unknown, and its
// children's Defaults are never evaluated. Setting this Default on the
// parent gives the object a known value up front so the per-field defaults
// still run for a partially-configured object (e.g. only `account_name`
// set).
type authenticationEnvDefault struct{}

func (authenticationEnvDefault) Description(_ context.Context) string {
	return "value defaults to the FASTLY_AZURE_ACCOUNT_NAME and FASTLY_AZURE_SHARED_ACCESS_SIGNATURE environment variables"
}

func (d authenticationEnvDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (authenticationEnvDefault) DefaultObject(ctx context.Context, _ fwdefaults.ObjectRequest, resp *fwdefaults.ObjectResponse) {
	resp.PlanValue = NewAuthenticationObject(
		envStringDefault(ctx, "FASTLY_AZURE_ACCOUNT_NAME"),
		envStringDefault(ctx, "FASTLY_AZURE_SHARED_ACCESS_SIGNATURE"),
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

func (n commonModel) AccountName() types.String {
	return authenticationValue(n.Authentication, "account_name")
}

func (n commonModel) SASToken() types.String {
	return authenticationValue(n.Authentication, "sas_token")
}

func (n commonModel) equal(other commonModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Container) == service.StringValue(other.Container) &&
		service.StringValue(n.AccountName()) == service.StringValue(other.AccountName()) &&
		service.StringValue(n.SASToken()) == service.StringValue(other.SASToken()) &&
		service.StringValue(n.Path) == service.StringValue(other.Path) &&
		service.Int64Value(n.Period) == service.Int64Value(other.Period) &&
		service.Int64Value(n.GzipLevel) == service.Int64Value(other.GzipLevel) &&
		service.StringValue(n.CompressionCodec) == service.StringValue(other.CompressionCodec) &&
		service.StringValue(n.MessageType) == service.StringValue(other.MessageType) &&
		service.StringValue(n.TimestampFormat) == service.StringValue(other.TimestampFormat) &&
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

// CommonAttributes returns the full Blob Storage logging attribute set — the
// shared attributes plus the VCL-only ones (format, format_version,
// placement, response_condition). Used by the standalone
// fastly_service_logging_blobstorage resource (which can attach to either
// service type) and the VCL nested block (NestedBlockSchema). Compute
// services use ComputeAttributes instead.
func CommonAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	maps.Copy(attrs, vclOnlyAttributes())
	return attrs
}

// ComputeAttributes returns the Blob Storage logging attribute set for
// Compute services, omitting the VCL-only attributes exposed by
// CommonAttributes.
func ComputeAttributes() map[string]schema.Attribute {
	return sharedAttributes()
}

// sharedAttributes returns the Blob Storage logging attributes common to
// both VCL and Compute services.
func sharedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		// Required
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name for the real-time logging configuration. Must be unique within the service.",
		},
		"container": schema.StringAttribute{
			Required:    true,
			Description: "The name of the Azure Blob Storage container in which to store logs.",
		},
		// Optional
		"authentication": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Default:     authenticationEnvDefault{},
			Description: "Azure authentication credentials for Blob Storage access. Both `account_name` and `sas_token` are required. When this block is omitted entirely, defaults to the `FASTLY_AZURE_ACCOUNT_NAME` and `FASTLY_AZURE_SHARED_ACCESS_SIGNATURE` environment variables.",
			Validators: []validator.Object{
				authenticationRequired{},
			},
			Attributes: map[string]schema.Attribute{
				"account_name": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     defaults.EnvString("FASTLY_AZURE_ACCOUNT_NAME", ""),
					Description: "The unique Azure Blob Storage namespace in which your data objects are stored. Can be set via the `FASTLY_AZURE_ACCOUNT_NAME` environment variable.",
				},
				"sas_token": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     defaults.EnvString("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", ""),
					Description: "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work. Can be set via the `FASTLY_AZURE_SHARED_ACCESS_SIGNATURE` environment variable.",
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
		"file_max_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultFileMaxBytes),
			Validators: []validator.Int64{
				fileMaxBytesRange{},
			},
			Description: "The maximum number of bytes for each uploaded file. A value of `0` can be used to indicate there is no limit on the size of uploaded files, otherwise the minimum value is `1048576` bytes (1 MiB).",
		},
		"gzip_level": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultGzipLevel),
			// compression_codec and gzip_level are mutually exclusive; the API
			// rejects a request that sets both. Validation runs against config,
			// where an unset gzip_level is null rather than the -1 default, so
			// this correctly fires only when both are set. int64validator.Between
			// likewise skips a null (omitted) config value, so it only rejects an
			// explicitly configured value outside 0-9 — including the internal -1
			// sentinel itself, which a user should never configure directly (omit
			// the attribute instead to get the same "unset" behavior).
			Validators: []validator.Int64{
				gzipLevelCodecConflict{},
				int64validator.Between(0, 9),
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
			Description: "How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `classic`.",
		},
		"path": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultPath),
			Description: "The path to upload logs to. Must end with a trailing slash. If this field is left empty, the files will be saved in the container's root path.",
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
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
		},
		"timestamp_format": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultTimestampFormat),
			Description: "`strftime`-specified timestamp format for log filename.",
		},
	}
}

// vclOnlyAttributes returns the Blob Storage logging attributes that only
// affect generated VCL and have no meaning for Compute services.
func vclOnlyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(constants.LoggingBlobStorageDefaultFormat),
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
			Description: "The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.",
		},
		"placement": schema.StringAttribute{
			// Not Computed: unset and explicitly "none" are distinct states on the
			// API (unset lets the API auto-place the logging call; "none" suppresses
			// it entirely). Keeping it plain Optional means removing placement from
			// config is a real, visible diff from Terraform core's own plan proposal,
			// with no plan modifier required to force it.
			//
			// On a VCL service the API resolves this to exactly what was configured.
			// Compute (wasm) services are the exception — the API forces "none" there
			// regardless of what was sent, which can never match the planned null, so
			// ResetVCLOnlyToDefaults discards it after apply for those services.
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("none"),
			},
			Description: "Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.",
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

// NestedBlockSchema returns the Blob Storage logging nested block schema for
// VCL services (service_cdn_auto.logging_blobstorage), including the
// VCL-only attributes.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Blob Storage logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ComputeNestedBlockSchema returns the Blob Storage logging nested block
// schema for Compute services (service_compute_auto.logging_blobstorage),
// omitting the VCL-only attributes.
func ComputeNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Blob Storage logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: ComputeAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.BlobStorage, error) {
	return client.ListBlobStorages(ctx, &fastly.ListBlobStoragesInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.BlobStorage) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteBlobStorage(ctx, &fastly.DeleteBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.BlobStorage, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateBlobStorage(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.BlobStorage) bool {
	remoteModel := FlattenToNestedModel(remote)
	preserveGzipSentinel(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.BlobStorage, error) {
	input := BuildUpdateInput(serviceID, version, desired)
	return client.UpdateBlobStorage(ctx, input)
}

func (o ops) ToModel(api *fastly.BlobStorage) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.BlobStorage]{
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

func (o computeOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.BlobStorage, error) {
	return client.ListBlobStorages(ctx, &fastly.ListBlobStoragesInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o computeOps) GetName(api *fastly.BlobStorage) string {
	return fastly.ToValue(api.Name)
}

func (o computeOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteBlobStorage(ctx, &fastly.DeleteBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o computeOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.BlobStorage, error) {
	input := BuildComputeCreateInput(serviceID, version, desired)
	return client.CreateBlobStorage(ctx, input)
}

func (o computeOps) Equal(desired ComputeNestedModel, remote *fastly.BlobStorage) bool {
	remoteModel := FlattenToComputeNestedModel(remote)
	preserveGzipSentinelCompute(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o computeOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.BlobStorage, error) {
	input := BuildComputeUpdateInput(serviceID, version, desired)
	return client.UpdateBlobStorage(ctx, input)
}

func (o computeOps) ToModel(api *fastly.BlobStorage) ComputeNestedModel {
	return FlattenToComputeNestedModel(api)
}

var computeReconciler = &reconcile.Resource[ComputeNestedModel, fastly.BlobStorage]{
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
