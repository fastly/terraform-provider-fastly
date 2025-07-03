---
page_title: Managing content in Fastly Object Storage
subcategory: "Guides"
---
# Managing content in Fastly Object Storage

Content in Fastly Object Storage (buckets and objects) can be managed
using a combination of this provider (Fastly) and the Amazon Web
Services (AWS) provider, since Fastly Object Storage provides an AWS
S3-compatible API.

## Example Configuration

Two HCL files are required. The first, `main.tf`:

```terraform
terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = "~> 7.1.0"
    }
  }
}

resource "fastly_object_storage_access_keys" "main" {
  description = "FOS Key"
  permission  = "read-write-admin"
}

module "fos" {
  source = "./fos"
  access_key_id = fastly_object_storage_access_keys.main.id
  secret_key    = fastly_object_storage_access_keys.main.secret_key
}
```

The second will need to be placed in a directory named `fos`, and
named `fos/main.tf`:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 6.0"
    }
  }
}

provider "aws" {
  access_key = var.access_key_id
  secret_key = var.secret_key
  region     = "us-east"

  s3_use_path_style                  = true
  skip_credentials_validation        = true
  skip_metadata_api_check            = true
  skip_region_validation             = true
  skip_requesting_account_id         = true

  endpoints {
    s3 = "https://us-east.object.fastlystorage.app"
  }
}

resource "aws_s3_bucket" "main" {
  bucket = "my-test-bucket-123"

  # Fastly object storage uses different region names
  lifecycle {
    ignore_changes = [region]
  }
}

```

Note: This example uses the `us-east` region of Fastly Object Storage;
if you wish to use a different region, ensure that the proper region
code is included in the `endpoints` block above.

## Getting Started

With the example files in place, you'll need to initialize Terraform
and obtain Fastly Object Storage credentials.

```bash
export FASTLY_API_KEY=<your Fastly API key here>
terraform init
terraform apply -target=fastly_object_storage_access_keys.main
```

Note: Terraform will issue a warning because the `-target` option is
used. This usage of that option is safe.

This step will connect to the Fastly API using the Fastly Terraform
provider and obtain a set of Fastly Object Storage credentials. Those
credentials will be stored in the Terraform state files (or other
state storage), they will not be displayed.

This initial step is necessary because the credentials are required by
the AWS Terraform provider, and if Terraform attempts to apply the
entire configuration the AWS provider will report an error because the
credentials are missing.

## Completing the Process

With the credentials obtained, a normal Terraform `apply` step can be
used to create the remaining infrastructure; in this case a bucket in
Fastly Object Storage named `my-test-bucket-123`. Terraform will pass
the Fastly Object Storage credentials to the AWS Terraform provider so
that it can use them to authenticate its API interactions with the
Fastly Object Storage system.

```bash
terraform apply
```
