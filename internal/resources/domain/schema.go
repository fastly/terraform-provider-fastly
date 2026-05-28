package domain

import (
	"context"
	"sort"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The domain that this service responds to.",
		},
		"comment": schema.StringAttribute{
			Optional:    true,
			Description: "Optional comment for the domain.",
		},
	}
}

func ResourceAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier.",
		},
		"service_id": schema.StringAttribute{
			Required:    true,
			Description: "Fastly service ID.",
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
		},
	}
	for k, v := range CommonAttributes() {
		attrs[k] = v
	}
	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Domains attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

func Normalize(input []NestedModel) []NestedModel {
	out := make([]NestedModel, len(input))
	copy(out, input)

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name.ValueString() < out[j].Name.ValueString()
	})

	return out
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	remote, err := client.ListDomains(ctx, &fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	result := make([]NestedModel, 0, len(remote))
	for _, d := range remote {
		model := NestedModel{
			Name: types.StringValue(fastly.ToValue(d.Name)),
		}
		if d.Comment != nil && *d.Comment != "" {
			model.Comment = types.StringValue(*d.Comment)
		} else {
			model.Comment = types.StringNull()
		}
		result = append(result, model)
	}

	return Normalize(result), nil
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	remote, err := client.ListDomains(ctx, &fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}

	desired = Normalize(desired)

	remoteByName := make(map[string]*fastly.Domain, len(remote))
	for _, d := range remote {
		remoteByName[fastly.ToValue(d.Name)] = d
	}

	desiredByName := make(map[string]NestedModel, len(desired))
	for _, d := range desired {
		desiredByName[d.Name.ValueString()] = d
	}

	// Delete domains no longer present.
	for name := range remoteByName {
		if _, ok := desiredByName[name]; !ok {
			err := client.DeleteDomain(ctx, &fastly.DeleteDomainInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
				Name:           name,
			})
			if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.StatusCode == 404 {
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	// Create or update desired domains.
	for _, desiredDomain := range desired {
		name := desiredDomain.Name.ValueString()
		remoteDomain, exists := remoteByName[name]

		comment := ""
		if !desiredDomain.Comment.IsNull() && !desiredDomain.Comment.IsUnknown() {
			comment = desiredDomain.Comment.ValueString()
		}

		if !exists {
			input := &fastly.CreateDomainInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
				Name:           fastly.ToPointer(name),
			}
			if !desiredDomain.Comment.IsUnknown() {
				input.Comment = fastly.ToPointer(comment)
			}
			if _, err := client.CreateDomain(ctx, input); err != nil {
				return err
			}
			continue
		}

		remoteComment := ""
		if remoteDomain.Comment != nil {
			remoteComment = *remoteDomain.Comment
		}

		if remoteComment != comment {
			input := &fastly.UpdateDomainInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
				Name:           name,
				Comment:        fastly.ToPointer(comment),
			}
			if _, err := client.UpdateDomain(ctx, input); err != nil {
				return err
			}
		}
	}

	return nil
}

func Equal(a, b []NestedModel) bool {
	a = Normalize(a)
	b = Normalize(b)

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Name.ValueString() != b[i].Name.ValueString() {
			return false
		}

		ac := ""
		if !a[i].Comment.IsNull() && !a[i].Comment.IsUnknown() {
			ac = a[i].Comment.ValueString()
		}

		bc := ""
		if !b[i].Comment.IsNull() && !b[i].Comment.IsUnknown() {
			bc = b[i].Comment.ValueString()
		}

		if ac != bc {
			return false
		}
	}

	return true
}
