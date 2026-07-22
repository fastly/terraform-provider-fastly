package loggings3

import (
	fastly "github.com/fastly/go-fastly/v16/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateS3Input {
	input := &fastly.CreateS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		BucketName:     new(service.StringValue(m.BucketName)),
	}

	input.AccessKey = fastly.NullString(service.StringValue(m.AccessKey()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.IAMRole = fastly.NullString(service.StringValue(m.IAMRole()))
	input.Domain = fastly.NullString(service.StringValue(m.Domain))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted. fastly.NullInt is
	// not used here because it treats 0 as unset too, which would silently drop
	// an explicit "no compression" (gzip_level = 0).
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = fastly.NullString(service.StringValue(m.ResponseCondition))
	input.PublicKey = fastly.NullString(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	if acl := service.StringValue(m.ACL); acl != "" {
		v := fastly.S3AccessControlList(acl)
		input.ACL = &v
	}
	if red := service.StringValue(m.Redundancy); red != "" {
		v := fastly.S3Redundancy(red)
		input.Redundancy = &v
	}
	if enc := service.StringValue(m.ServerSideEncryption); enc != "" {
		v := fastly.S3ServerSideEncryption(enc)
		input.ServerSideEncryption = &v
	}
	input.ServerSideEncryptionKMSKeyID = fastly.NullString(service.StringValue(m.ServerSideEncryptionKMSKeyID))

	if fmb := service.Int64Value(m.FileMaxBytes); fmb != 0 {
		v := int(fmb)
		input.FileMaxBytes = &v
	}

	return input
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateS3Input. The standalone
// fastly_service_logging_s3 resource shares one schema across both service
// types, so this is called instead of BuildComputeCreateInput to strip the
// VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateS3Input) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for UpdateS3Input.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateS3Input) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// BuildComputeCreateInput is BuildCreateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateS3Input {
	input := &fastly.CreateS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		BucketName:     new(service.StringValue(m.BucketName)),
	}

	input.AccessKey = fastly.NullString(service.StringValue(m.AccessKey()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.IAMRole = fastly.NullString(service.StringValue(m.IAMRole()))
	input.Domain = fastly.NullString(service.StringValue(m.Domain))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted. fastly.NullInt is
	// not used here because it treats 0 as unset too, which would silently drop
	// an explicit "no compression" (gzip_level = 0).
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.PublicKey = fastly.NullString(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	if acl := service.StringValue(m.ACL); acl != "" {
		v := fastly.S3AccessControlList(acl)
		input.ACL = &v
	}
	if red := service.StringValue(m.Redundancy); red != "" {
		v := fastly.S3Redundancy(red)
		input.Redundancy = &v
	}
	if enc := service.StringValue(m.ServerSideEncryption); enc != "" {
		v := fastly.S3ServerSideEncryption(enc)
		input.ServerSideEncryption = &v
	}
	input.ServerSideEncryptionKMSKeyID = fastly.NullString(service.StringValue(m.ServerSideEncryptionKMSKeyID))

	if fmb := service.Int64Value(m.FileMaxBytes); fmb != 0 {
		v := int(fmb)
		input.FileMaxBytes = &v
	}

	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateS3Input {
	input := &fastly.UpdateS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		BucketName:     new(service.StringValue(m.BucketName)),
	}

	input.AccessKey = fastly.NullString(service.StringValue(m.AccessKey()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.IAMRole = fastly.NullString(service.StringValue(m.IAMRole()))
	input.Domain = fastly.NullString(service.StringValue(m.Domain))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted.
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.PublicKey = fastly.NullString(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	aclVal := fastly.S3AccessControlList(service.StringValue(m.ACL))
	input.ACL = &aclVal

	redVal := fastly.S3Redundancy(service.StringValue(m.Redundancy))
	input.Redundancy = &redVal

	if enc := service.StringValue(m.ServerSideEncryption); enc != "" {
		v := fastly.S3ServerSideEncryption(enc)
		input.ServerSideEncryption = &v
	}
	input.ServerSideEncryptionKMSKeyID = fastly.NullString(service.StringValue(m.ServerSideEncryptionKMSKeyID))

	fmb := int(service.Int64Value(m.FileMaxBytes))
	input.FileMaxBytes = &fmb

	return input
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateS3Input {
	input := &fastly.UpdateS3Input{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		BucketName:     new(service.StringValue(m.BucketName)),
	}

	input.AccessKey = fastly.NullString(service.StringValue(m.AccessKey()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.IAMRole = fastly.NullString(service.StringValue(m.IAMRole()))
	input.Domain = fastly.NullString(service.StringValue(m.Domain))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted.
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = fastly.NullString(service.StringValue(m.ResponseCondition))
	input.PublicKey = fastly.NullString(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	aclVal := fastly.S3AccessControlList(service.StringValue(m.ACL))
	input.ACL = &aclVal

	redVal := fastly.S3Redundancy(service.StringValue(m.Redundancy))
	input.Redundancy = &redVal

	if enc := service.StringValue(m.ServerSideEncryption); enc != "" {
		v := fastly.S3ServerSideEncryption(enc)
		input.ServerSideEncryption = &v
	}
	input.ServerSideEncryptionKMSKeyID = fastly.NullString(service.StringValue(m.ServerSideEncryptionKMSKeyID))

	fmb := int(service.Int64Value(m.FileMaxBytes))
	input.FileMaxBytes = &fmb

	return input
}
