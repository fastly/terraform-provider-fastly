package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var computeAttributes = ServiceMetadata{
	ServiceTypeCompute,
}

// Ordering is important - stored is processing order
// Some objects may need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc).
var computeService = &BaseServiceDefinition{
	Type: computeAttributes.serviceType,
	Attributes: []ServiceAttributeDefinition{
		NewServiceProductEnablement(computeAttributes),
		NewServiceDomain(computeAttributes),
		NewServiceBackend(computeAttributes),
		NewServiceLoggingS3(computeAttributes),
		NewServiceLoggingPaperTrail(computeAttributes),
		NewServiceLoggingSumologic(computeAttributes),
		NewServiceLoggingGCS(computeAttributes),
		NewServiceLoggingBigQuery(computeAttributes),
		NewServiceLoggingSyslog(computeAttributes),
		NewServiceLoggingLogentries(computeAttributes),
		NewServiceLoggingSplunk(computeAttributes),
		NewServiceLoggingBlobStorage(computeAttributes),
		NewServiceLoggingHTTPS(computeAttributes),
		NewServiceLoggingElasticSearch(computeAttributes),
		NewServiceLoggingFTP(computeAttributes),
		NewServiceLoggingSFTP(computeAttributes),
		NewServiceLoggingDatadog(computeAttributes),
		NewServiceLoggingLoggly(computeAttributes),
		NewServiceLoggingGooglePubSub(computeAttributes),
		NewServiceLoggingScalyr(computeAttributes),
		NewServiceLoggingNewRelic(computeAttributes),
		NewServiceLoggingKafka(computeAttributes),
		NewServiceLoggingHeroku(computeAttributes),
		NewServiceLoggingHoneycomb(computeAttributes),
		NewServiceLoggingLogshuttle(computeAttributes),
		NewServiceLoggingOpenstack(computeAttributes),
		NewServiceLoggingDigitalOcean(computeAttributes),
		NewServiceLoggingCloudfiles(computeAttributes),
		NewServiceLoggingKinesis(computeAttributes),
		NewServiceDictionary(computeAttributes),
		NewServicePackage(computeAttributes),
	},
}

func resourceServiceCompute() *schema.Resource {
	return resourceService(computeService)
}
