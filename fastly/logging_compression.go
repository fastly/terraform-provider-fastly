package fastly

import (
	"fmt"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// Valid compression options for logging endpoints.
const (
	CompressionZstd   = "zstd"
	CompressionSnappy = "snappy"
	CompressionGzip0  = "gzip-0"
	CompressionGzip1  = "gzip-1"
	CompressionGzip2  = "gzip-2"
	CompressionGzip3  = "gzip-3"
	CompressionGzip4  = "gzip-4"
	CompressionGzip5  = "gzip-5"
	CompressionGzip6  = "gzip-6"
	CompressionGzip7  = "gzip-7"
	CompressionGzip8  = "gzip-8"
	CompressionGzip9  = "gzip-9"
	CompressionNone   = "none"
)

// CompressionDescription describes the compression field for logging endpoints.
const CompressionDescription = "Compression format for log data. Valid values: zstd, snappy, gzip-0 through gzip-9, or none"

// LoggingCompressionOptions returns all valid compression options.
func LoggingCompressionOptions() []string {
	return []string{
		CompressionZstd,
		CompressionSnappy,
		CompressionNone,
		CompressionGzip0,
		CompressionGzip1,
		CompressionGzip2,
		CompressionGzip3,
		CompressionGzip4,
		CompressionGzip5,
		CompressionGzip6,
		CompressionGzip7,
		CompressionGzip8,
		CompressionGzip9,
	}
}

// CompressionToAPIFields converts the user-friendly compression field value
// to the API's compression_codec and gzip_level fields.
//
// Mapping:
// - "zstd"   -> compression_codec: "zstd", gzip_level: nil
// - "snappy" -> compression_codec: "snappy", gzip_level: nil
// - "none"   -> compression_codec: "" (empty string), gzip_level: 0
// - "gzip-0" -> compression_codec: "gzip", gzip_level: 0
// - "gzip-1" through "gzip-9" -> (nil), gzip_level: 1-9
//
// Note: When transitioning from zstd/snappy to none or gzip levels, we must
// explicitly send compression_codec as empty string ("") to clear the previous value.
func CompressionToAPIFields(compression string) (compressionCodec *string, gzipLevel *int) {
	if compression == "" {
		return nil, nil
	}
	switch compression {
	case CompressionZstd:
		return gofastly.ToPointer("zstd"), nil
	case CompressionSnappy:
		return gofastly.ToPointer("snappy"), nil
	case CompressionNone:
		return gofastly.ToPointer(""), gofastly.ToPointer(0)
	case CompressionGzip0:
		return gofastly.ToPointer("gzip"), gofastly.ToPointer(0)
	case CompressionGzip1:
		return nil, gofastly.ToPointer(1)
	case CompressionGzip2:
		return nil, gofastly.ToPointer(2)
	case CompressionGzip3:
		return nil, gofastly.ToPointer(3)
	case CompressionGzip4:
		return nil, gofastly.ToPointer(4)
	case CompressionGzip5:
		return nil, gofastly.ToPointer(5)
	case CompressionGzip6:
		return nil, gofastly.ToPointer(6)
	case CompressionGzip7:
		return nil, gofastly.ToPointer(7)
	case CompressionGzip8:
		return nil, gofastly.ToPointer(8)
	case CompressionGzip9:
		return nil, gofastly.ToPointer(9)
	default:
		return nil, nil
	}
}

// APIFieldsToCompression converts the API's compression_codec and gzip_level
// fields to the user-friendly compression field value.
//
// Reverse mapping:
// - compression_codec: "zstd" -> "zstd"
// - compression_codec: "snappy" -> "snappy"
// - compression_codec: "" (empty string), gzip_level: 0 -> "none"
// - compression_codec: "gzip", gzip_level: 0 -> "gzip-0"
// - compression_codec: "gzip", gzip_level: nil -> "gzip-3" (default)
// - compression_codec: nil, gzip_level: 1-9 -> "gzip-1" through "gzip-9"
// - compression_codec: nil, gzip_level: 0 -> "" (not set by user).
func APIFieldsToCompression(compressionCodec *string, gzipLevel *int) string {
	if compressionCodec != nil {
		if *compressionCodec != "" {
			switch *compressionCodec {
			case "zstd":
				return CompressionZstd
			case "snappy":
				return CompressionSnappy
			case "gzip":
				if gzipLevel != nil {
					return fmt.Sprintf("gzip-%d", *gzipLevel)
				}
				return CompressionGzip3 // Default gzip level when codec is gzip but not specified
			}
		} else if *compressionCodec == "" && gzipLevel != nil && *gzipLevel == 0 {
			// When compression_codec is explicitly "" (empty string) and gzip_level is 0,
			// this means the user set compression = "none"
			return CompressionNone
		}
	}

	// If compressionCodec is nil and gzipLevel is 0 or nil, compression was not set by the user
	// Return empty string so it won't be added to state
	if gzipLevel != nil && *gzipLevel >= 1 && *gzipLevel <= 9 {
		return fmt.Sprintf("gzip-%d", *gzipLevel)
	}
	return ""
}
