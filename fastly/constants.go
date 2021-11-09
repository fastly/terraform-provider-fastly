package fastly

const MessageTypeDescription = "How the message should be formatted. Can be either `classic`, `loggly`, `logplex` or `blank`. Default is `classic`"
const GzipLevelDescription = "Level of Gzip compression from `0-9`. `0` means no compression. `1` is the fastest and the least compressed version, `9` is the slowest and the most compressed version. Default `0`"
const TimestampFormatDescription = "The `strftime` specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)"
