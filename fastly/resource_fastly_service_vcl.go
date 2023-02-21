package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var vclAttributes = ServiceMetadata{
	ServiceTypeVCL,
}

// Ordering is important - stored is processing order
// Conditions need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc)
var vclService = &BaseServiceDefinition{
	Type: vclAttributes.serviceType,
	Attributes: []ServiceAttributeDefinition{
		NewServiceSettings(),
		NewServiceCondition(vclAttributes),
		NewServiceDomain(vclAttributes),
		NewServiceHealthCheck(vclAttributes),
		NewServiceBackend(vclAttributes),
		NewServiceProductEnablement(vclAttributes),
		NewServiceDirector(vclAttributes),
		NewServiceHeader(vclAttributes),
		NewServiceGzip(vclAttributes),
		NewServiceLoggingS3(vclAttributes),
		NewServiceLoggingPaperTrail(vclAttributes),
		NewServiceLoggingSumologic(vclAttributes),
		NewServiceLoggingGCS(vclAttributes),
		NewServiceLoggingBigQuery(vclAttributes),
		NewServiceLoggingSyslog(vclAttributes),
		NewServiceLoggingLogentries(vclAttributes),
		NewServiceLoggingSplunk(vclAttributes),
		NewServiceLoggingBlobStorage(vclAttributes),
		NewServiceLoggingHTTPS(vclAttributes),
		NewServiceLoggingElasticSearch(vclAttributes),
		NewServiceLoggingFTP(vclAttributes),
		NewServiceLoggingSFTP(vclAttributes),
		NewServiceLoggingDatadog(vclAttributes),
		NewServiceLoggingLoggly(vclAttributes),
		NewServiceLoggingGooglePubSub(vclAttributes),
		NewServiceLoggingScalyr(vclAttributes),
		NewServiceLoggingNewRelic(vclAttributes),
		NewServiceLoggingKafka(vclAttributes),
		NewServiceLoggingHeroku(vclAttributes),
		NewServiceLoggingHoneycomb(vclAttributes),
		NewServiceLoggingLogshuttle(vclAttributes),
		NewServiceLoggingOpenstack(vclAttributes),
		NewServiceLoggingDigitalOcean(vclAttributes),
		NewServiceLoggingCloudfiles(vclAttributes),
		NewServiceLoggingKinesis(vclAttributes),
		NewServiceResponseObject(vclAttributes),
		NewServiceRequestSetting(vclAttributes),
		NewServiceVCL(vclAttributes),
		NewServiceSnippet(vclAttributes),
		NewServiceDynamicSnippet(vclAttributes),
		NewServiceCacheSetting(vclAttributes),
		NewServiceACL(),
		NewServiceDictionary(vclAttributes),
		NewServiceWAF(vclAttributes),
	},
}

func resourceServiceVCL() *schema.Resource {
	return resourceService(vclService)
}
