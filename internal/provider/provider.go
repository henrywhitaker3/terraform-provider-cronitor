// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &CronitorProvider{}
var _ provider.ProviderWithFunctions = &CronitorProvider{}

// ScaffoldingProvider defines the provider implementation.
type CronitorProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type CronitorProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiKey   types.String `tfsdk:"api_key"`
}

func (p *CronitorProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cronitor"
	resp.Version = p.version
}

func (p *CronitorProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The api key used to connect to cronitor",
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The cronitor base API endpoint",
				Optional:            true,
			},
		},
	}
}

func (p *CronitorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CronitorProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := ""
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.String()
	}

	// Example client configuration for data sources and resources
	client := cronitor.NewClient(cronitor.NewClientOpts{
		ApiKey:   data.ApiKey.ValueString(),
		Endpoint: endpoint,
	})
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CronitorProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentResource,
		NewHeartbeatMonitorResource,
		NewHttpMonitorResource,
		NewNotificationListResource,
	}
}

func (p *CronitorProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *CronitorProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CronitorProvider{
			version: version,
		}
	}
}
