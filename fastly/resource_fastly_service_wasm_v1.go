package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Ordering is important - stored is processing order
// Conditions need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc)

var wasmService = &BaseServiceDefinition{
	Type: "wasm",
	Attributes: []ServiceAttributeDefinition{
		NewServiceDomain(),
		NewServiceHealthCheck(),
		NewServiceBackend(),
		//These have been commented out until they are satisfactorily integrated with Wasm resources
		//NewServiceS3Logging(),
		//NewServicePaperTrail(),
		//NewServiceSumologic(),
		//NewServiceGCSLogging(),
		//NewServiceBigQueryLogging(),
		//NewServiceSyslog(),
		//NewServiceLogentries(),
		//NewServiceSplunk(),
		//NewServiceBlobStorageLogging(),
		//NewServiceHTTPSLogging(),
		//NewServiceLoggingElasticSearch(),
		//NewServiceLoggingFTP(),
		//NewServiceLoggingSFTP(),
		//NewServiceLoggingDatadog(),
		//NewServiceLoggingLoggly(),
		//NewServiceLoggingGooglePubSub(),
		//NewServiceLoggingScalyr(),
		//NewServiceLoggingNewRelic(),
		//NewServiceLoggingKafka(),
		NewServicePackage(),
	},
}

func resourceServiceWasmV1() *schema.Resource {
	return resourceService(wasmService)
}
