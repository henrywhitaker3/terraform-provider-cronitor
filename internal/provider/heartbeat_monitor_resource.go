// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HeartbeatMonitorResource{}
var _ resource.ResourceWithImportState = &HeartbeatMonitorResource{}

func NewHeartbeatMonitorResource() resource.Resource {
	return &HeartbeatMonitorResource{}
}

// HeartbeatMonitorResource defines the resource implementation.
type HeartbeatMonitorResource struct {
	client *cronitor.Client
}

func (r *HeartbeatMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heartbeat_monitor"
}

func (r *HeartbeatMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Heartbeat Monitor resource",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "The monitor id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The monitor name",
				Required:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the monitor is disabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"failure_tolerance": schema.Int32Attribute{
				MarkdownDescription: "The number of times the monitor can fail before triggering an alert",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
			},
			"grace_seconds": schema.Int32Attribute{
				MarkdownDescription: "The number of seconds to wait after failure before triggering an alert",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the monitor is paused",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"realert_interval": schema.StringAttribute{
				MarkdownDescription: "The interval that alerts are re-sent at",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("every 8 hours"),
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "The schedule the monitor runs on",
				Required:            true,
			},
			"schedule_tolerance": schema.Int32Attribute{
				MarkdownDescription: "The number of missed scheduled executions before triggering an alert",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The monitor tags",
				Optional:            true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "The timezone of the schedule",
				Optional:            true,
			},
			"notify": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Where the alerts are sent when a failure occurs",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default")})),
			},
			"environments": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The environments the monitor runs in",
				Optional:            true,
			},
			"telemetry_url": schema.StringAttribute{
				MarkdownDescription: "The url to send pings to",
				Sensitive:           true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The group the monitor belongs to",
				Optional:            true,
			},
		},
	}
}

func (r *HeartbeatMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HeartbeatMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HeartbeatMonitorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := r.client.CreateMonitor(ctx, heartbeatToMonitorRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("failed to create monitor", err.Error())
		return
	}

	data.Key = types.StringValue(*monitor.Key)
	data.TelemetryUrl = types.StringValue(fmt.Sprintf("https://cronitor.link/p/%s/%s", r.client.ApiKey, *monitor.Key))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HeartbeatMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HeartbeatMonitorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := heartbeatToMonitorRequest(data)

	monitor, err := r.client.GetMonitor(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get monitor from api", err.Error())
		return
	}

	fixSliceOrder(state.Assertions, &monitor.Assertions)
	fixSliceOrder(state.Environments, &monitor.Environments)
	fixSliceOrder(state.Tags, &monitor.Tags)

	data = toHeartbeatMonitor(monitor)
	data.TelemetryUrl = types.StringValue(fmt.Sprintf("https://cronitor.link/p/%s/%s", r.client.ApiKey, *monitor.Key))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HeartbeatMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state HeartbeatMonitorModel
	var plan HeartbeatMonitorModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	upd := heartbeatToMonitorRequest(plan)
	upd.Key = state.Key.ValueStringPointer()
	monitor, err := r.client.UpdateMonitor(ctx, upd)
	if err != nil {
		resp.Diagnostics.AddError("failed to update heartbeat monitor", err.Error())
		return
	}

	fixSliceOrder(upd.Assertions, &monitor.Assertions)
	fixSliceOrder(upd.Environments, &monitor.Environments)
	fixSliceOrder(upd.Tags, &monitor.Tags)

	state = toHeartbeatMonitor(monitor)
	state.TelemetryUrl = types.StringValue(fmt.Sprintf("https://cronitor.link/p/%s/%s", r.client.ApiKey, *monitor.Key))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *HeartbeatMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HeartbeatMonitorModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMonitor(ctx, data.Key.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete record", err.Error())
		return
	}
}

func (r *HeartbeatMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func (r *HeartbeatMonitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data HeartbeatMonitorModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if err := data.validate(); err != nil {
	// 	resp.Diagnostics.AddError("monitor failed validation", err.Error())
	// 	return
	// }
}
