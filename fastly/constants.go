package fastly

// MessageTypeDescription describes the message format.
const MessageTypeDescription = "How the message should be formatted. Can be either `classic`, `loggly`, `logplex` or `blank`. Default is `classic`"

// GzipLevelDescription describes Gzip compression.
const GzipLevelDescription = "Level of Gzip compression from `0-9`. `0` means no compression. `1` is the fastest and the least compressed version, `9` is the slowest and the most compressed version. Default `0`"

// TimestampFormatDescription describes the timestamp format.
const TimestampFormatDescription = "The `strftime` specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)"

// SnippetTypeDescription describes the VCL snippet location.
const SnippetTypeDescription = "The location in generated VCL where the snippet should be placed (can be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log` or `none`)"
