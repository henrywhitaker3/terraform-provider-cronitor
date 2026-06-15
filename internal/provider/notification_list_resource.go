// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NotificationListResource{}
var _ resource.ResourceWithImportState = &NotificationListResource{}

func NewNotificationListResource() resource.Resource {
	return &NotificationListResource{}
}

// NotificationListResource defines the resource implementation.
type NotificationListResource struct {
	client *cronitor.Client
}

func (r *NotificationListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_list"
}

func (r *NotificationListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Notification List resource",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "The notification list id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The notification list name",
				Required:            true,
			},
			"emails": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The emails to send notifications to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"slack": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The slack channels to send notifications to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"pagerduty": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The slack channels to send notifications to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"phones": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The phone numbers to send notifications to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"webhooks": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The webhook urls to send notifications to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"environments": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The environments this notification list applies to",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
		},
	}
}

func (r *NotificationListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cronitor.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cronitor.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *NotificationListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationListModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.client.CreateNotificationList(ctx, listToListRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification list", err.Error())
		return
	}

	data = toNotificationList(list)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationListModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := listToListRequest(data)

	list, err := r.client.GetNotificationList(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get notification list from api", err.Error())
		return
	}

	fixSliceOrder(state.Notifications.Emails, &list.Notifications.Emails)
	fixSliceOrder(state.Notifications.Slack, &list.Notifications.Slack)
	fixSliceOrder(state.Notifications.Pagerduty, &list.Notifications.Pagerduty)
	fixSliceOrder(state.Notifications.Phones, &list.Notifications.Phones)
	fixSliceOrder(state.Notifications.Webhooks, &list.Notifications.Webhooks)
	fixSliceOrder(state.Environments, &list.Environments)

	data = toNotificationList(list)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state NotificationListModel
	var plan NotificationListModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	upd := listToListRequest(plan)
	list, err := r.client.UpdateNotificationList(ctx, upd)
	if err != nil {
		resp.Diagnostics.AddError("failed to update heartbeat monitor", err.Error())
		return
	}

	fixSliceOrder(upd.Notifications.Emails, &list.Notifications.Emails)
	fixSliceOrder(upd.Notifications.Slack, &list.Notifications.Slack)
	fixSliceOrder(upd.Notifications.Pagerduty, &list.Notifications.Pagerduty)
	fixSliceOrder(upd.Notifications.Phones, &list.Notifications.Phones)
	fixSliceOrder(upd.Notifications.Webhooks, &list.Notifications.Webhooks)
	fixSliceOrder(upd.Environments, &list.Environments)

	state = toNotificationList(list)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationListModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNotificationList(ctx, listToListRequest(data)); err != nil {
		resp.Diagnostics.AddError("failed to delete record", err.Error())
		return
	}
}

func (r *NotificationListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func (r *NotificationListResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data NotificationListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if err := data.validate(); err != nil {
	// 	resp.Diagnostics.AddError("monitor failed validation", err.Error())
	// 	return
	// }
}
