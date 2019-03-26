package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceFastlyFlattenBlobStorage(t *testing.T) {
	cases := []struct {
		remote []*gofastly.BlobStorage
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.BlobStorage{
				{
					Name:              "test-blobstorage",
					Path:              "/logs/",
					AccountName:       "test",
					Container:         "fastly",
					SASToken:          "test-sas-token",
					Period:            12,
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
					GzipLevel:         9,
					PublicKey:         "test-public-key",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     1,
					MessageType:       "classic",
					Placement:         "waf_debug",
					ResponseCondition: "error_response",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "test-blobstorage",
					"path":               "/logs/",
					"account_name":       "test",
					"container":          "fastly",
					"sas_token":          "test-sas-token",
					"period":             uint(12),
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"gzip_level":         uint(9),
					"public_key":         "test-public-key",
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(1),
					"message_type":       "classic",
					"placement":          "waf_debug",
					"response_condition": "error_response",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBlobStorages(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_blobstoragelogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	blobStorageLogOne := gofastly.BlobStorage{
		Name:              "test-blobstorage-1",
		Path:              "/5XX/",
		AccountName:       "test",
		Container:         "fastly",
		SASToken:          "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D",
		Period:            12,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         9,
		PublicKey:         "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n",
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		MessageType:       "blank",
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	blobStorageLogOneUpdated := gofastly.BlobStorage{
		Name:              "test-blobstorage-1",
		Path:              "/5XX/",
		AccountName:       "test",
		Container:         "fastly",
		SASToken:          "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D",
		Period:            12,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         9,
		PublicKey:         "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		MessageType:       "blank",
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	blobStorageLogTwo := gofastly.BlobStorage{
		Name:              "test-blobstorage-2",
		Path:              "/2XX/",
		AccountName:       "test",
		Container:         "fastly",
		SASToken:          "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D",
		Period:            12,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         9,
		PublicKey:         "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		MessageType:       "blank",
		Placement:         "waf_debug",
		ResponseCondition: "ok_response_2XX",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1BlobStorageLoggingConfig_complete(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1BlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLogOne}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "blobstoragelogging.#", "1"),
				),
			},

			{
				Config: testAccServiceV1BlobStorageLoggingConfig_update(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1BlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLogOneUpdated, &blobStorageLogTwo}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "blobstoragelogging.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_blobstoragelogging_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	blobStorageLog := gofastly.BlobStorage{
		Name:            "test-blobstorage",
		AccountName:     "test",
		Container:       "fastly",
		SASToken:        "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D",
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:       0,
		Format:          "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:   2,
		MessageType:     "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1BlobStorageLoggingConfig_default(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1BlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLog}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "blobstoragelogging.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_blobstoragelogging_env(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	// set env variable to something we expect
	resetEnv := setBlobStorageEnv("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D", t)
	defer resetEnv()

	blobStorageLog := gofastly.BlobStorage{
		Name:            "test-blobstorage",
		AccountName:     "test",
		Container:       "fastly",
		SASToken:        "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D",
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:       0,
		Format:          "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:   2,
		MessageType:     "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1BlobStorageLoggingConfig_env(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1BlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLog}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "blobstoragelogging.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1BlobStorageLoggingAttributes(service *gofastly.ServiceDetail, localBlobStorageList []*gofastly.BlobStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		remoteBlobStorageList, err := conn.ListBlobStorages(&gofastly.ListBlobStoragesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Blob Storage Logging for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(remoteBlobStorageList) != len(localBlobStorageList) {
			return fmt.Errorf("Blob Storage List count mismatch, expected (%d), got (%d)", len(localBlobStorageList), len(remoteBlobStorageList))
		}

		var found int
		for _, lbs := range localBlobStorageList {
			for _, rbs := range remoteBlobStorageList {
				if lbs.Name == rbs.Name {
					// we don't know these things ahead of time, so populate them now
					lbs.ServiceID = service.ID
					lbs.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					rbs.CreatedAt = nil
					rbs.UpdatedAt = nil
					if !reflect.DeepEqual(lbs, rbs) {
						return fmt.Errorf("Bad match Blob Storage logging match, expected (%#v), got (%#v)", lbs, rbs)
					}
					found++
				}
			}
		}

		if found != len(localBlobStorageList) {
			return fmt.Errorf("Error matching Blob Storage Logging rules")
		}

		return nil
	}
}

func testAccServiceV1BlobStorageLoggingConfig_complete(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  blobstoragelogging {
    name               = "test-blobstorage-1"
    path               = "/5XX/"
    account_name       = "test"
    container          = "fastly"
    sas_token          = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period             = 12
    timestamp_format   = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level         = 9
    public_key         = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n"
    format             = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    format_version     = 1
    message_type       = "blank"
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1BlobStorageLoggingConfig_update(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  condition {
    name      = "ok_response_2XX"
    statement = "resp.status >= 200 && resp.status < 300"
    priority  = 10
    type      = "RESPONSE"
  }

  blobstoragelogging {
    name               = "test-blobstorage-1"
    path               = "/5XX/"
    account_name       = "test"
    container          = "fastly"
    sas_token          = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period             = 12
    timestamp_format   = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level         = 9
    public_key         = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n"
    format             = "%%h %%l %%u %%{now}V %%{req.method}V %%{req.url}V %%>s %%{resp.http.Content-Length}V"
    format_version     = 2
    message_type       = "blank"
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  blobstoragelogging {
    name               = "test-blobstorage-2"
    path               = "/2XX/"
    account_name       = "test"
    container          = "fastly"
    sas_token          = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period             = 12
    timestamp_format   = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level         = 9
    public_key         = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/\nibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4\n8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p\nlDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn\ndwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB\n89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz\ndCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6\nvFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc\n9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9\nOLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX\nSvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq\n7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx\nkATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG\nM1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe\nu6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L\n4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF\nftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K\nUEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu\nYrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi\nkiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb\nDAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml\ndYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L\n3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c\nFaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR\n5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR\nwMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N\n=28dr\n-----END PGP PUBLIC KEY BLOCK-----\n"
    format             = "%%h %%l %%u %%{now}V %%{req.method}V %%{req.url}V %%>s %%{resp.http.Content-Length}V"
    format_version     = 2
    message_type       = "blank"
    placement          = "waf_debug"
    response_condition = "ok_response_2XX"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1BlobStorageLoggingConfig_default(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  blobstoragelogging {
    name         = "test-blobstorage"
    account_name = "test"
    container    = "fastly"
    sas_token    = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1BlobStorageLoggingConfig_env(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  blobstoragelogging {
    name         = "test-blobstorage"
    account_name = "test"
    container    = "fastly"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func setBlobStorageEnv(sas string, t *testing.T) func() {
	e := getBlobStorageEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", sas); err != nil {
		t.Fatalf("Error setting env var FASTLY_AZURE_SHARED_ACCESS_SIGNATURE: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", e.SASToken); err != nil {
			t.Fatalf("Error resetting env var FASTLY_AZURE_SHARED_ACCESS_SIGNATURE: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentBlobStorageEnv struct {
	SASToken string
}

func getBlobStorageEnv() *currentBlobStorageEnv {
	// Grab the existing Fastly Azure SAS token and preserve, in the off chance
	// they're actually set in the enviornment
	return &currentBlobStorageEnv{
		SASToken: os.Getenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE"),
	}
}
