package fastly

import (
	"testing"
)

func TestDataSourceFastlyDomainsV1_DeprecationRegistered(t *testing.T) {
	p := Provider()

	dataSource, ok := p.DataSourcesMap["fastly_domains_v1"]
	if !ok {
		t.Fatal("expected data source fastly_domains_v1 to be registered")
	}

	if dataSource.DeprecationMessage == "" {
		t.Fatal("expected fastly_domains_v1 to have a DeprecationMessage")
	}
}
