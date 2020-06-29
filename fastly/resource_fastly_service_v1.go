package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Ordering is important - stored is processing order
// Conditions need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc)
var vclService = &BaseServiceDefinition{
	Type: "vcl",
	Attributes: []ServiceAttributeDefinition{
		NewServiceSettings(),
		NewServiceCondition(),
		NewServiceDomain(),
		NewServiceHealthCheck(),
		NewServiceBackend(),
		NewServiceDirector(),
		NewServiceHeader(),
		NewServiceGZIP(),
		NewServiceS3Logging(),
		NewServicePaperTrail(),
		NewServiceSumologic(),
		NewServiceGCSLogging(),
		NewServiceBigQueryLogging(),
		NewServiceSyslog(),
		NewServiceLogentries(),
		NewServiceSplunk(),
		NewServiceBlobStorageLogging(),
		NewServiceHTTPSLogging(),
		NewServiceLoggingElasticSearch(),
		NewServiceLoggingFTP(),
		NewServiceLoggingSFTP(),
		NewServiceLoggingDatadog(),
		NewServiceLoggingLoggly(),
		NewServiceLoggingGooglePubSub(),
		NewServiceLoggingScalyr(),
		NewServiceLoggingNewRelic(),
		NewServiceLoggingKafka(),
		NewServiceLoggingHeroku(),
		NewServiceLoggingHoneycomb(),
		NewServiceLoggingLogshuttle(),
		NewServiceLoggingOpenstack(),
		NewServiceResponseObject(),
		NewServiceRequestSetting(),
		NewServiceVCL(),
		NewServiceSnippet(),
		NewServiceDynamicSnippet(),
		NewServiceCacheSetting(),
		NewServiceACL(),
		NewServiceDictionary(),
	},
}

func resourceServiceV1() *schema.Resource {
	return resourceService(vclService)
}
