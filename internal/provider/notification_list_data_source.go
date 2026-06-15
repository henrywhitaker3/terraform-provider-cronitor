// Copyright (c) Henry Whitaker
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NotificationListDataSource{}

func NewExampleDataSource() datasource.DataSource {
	return &NotificationListDataSource{}
}

// ExampleDataSource defines the data source implementation.
type NotificationListDataSource struct {
	client *cronitor.Client
}

func (n *NotificationListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_list"
}

func (n *NotificationListDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Notification list data source",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "The notification list id",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The notification list name",
				Computed:            true,
			},
			"emails": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The emails to send notifications to",
				Computed:            true,
			},
			"slack": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The slack channels to send notifications to",
				Computed:            true,
			},
			"pagerduty": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The slack channels to send notifications to",
				Computed:            true,
			},
			"phones": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The phone numbers to send notifications to",
				Computed:            true,
			},
			"webhooks": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The webhook urls to send notifications to",
				Computed:            true,
			},
			"environments": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The environments this notification list applies to",
				Computed:            true,
			},
		},
	}
}

func (n *NotificationListDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cronitor.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *cronitor.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	n.client = client
}

func (d *NotificationListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationListModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	list, err := d.client.GetNotificationList(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get notification list", err.Error())
		return
	}

	data = toNotificationList(list)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a notification list")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
