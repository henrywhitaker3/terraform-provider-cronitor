package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

var _ resource.Resource = &EnvironmentResource{}
var _ resource.ResourceWithImportState = &EnvironmentResource{}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

type EnvironmentResource struct {
	client *cronitor.Client
}

func (r *EnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Environment resource",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "The environment key used as its identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The environment display name",
				Required:            true,
			},
			"with_alerts": schema.BoolAttribute{
				MarkdownDescription: "Whether alerts are enabled for this environment",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *EnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.CreateEnvironment(ctx, &cronitor.Environment{
		Key:        data.Key.ValueString(),
		Name:       data.Name.ValueString(),
		WithAlerts: data.WithAlerts.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create environment", err.Error())
		return
	}

	data.Key = types.StringValue(env.Key)
	data.Name = types.StringValue(env.Name)
	data.WithAlerts = types.BoolValue(env.WithAlerts)

	tflog.Trace(ctx, "created an environment")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnvironment(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get environment from api", err.Error())
		return
	}

	data.Key = types.StringValue(env.Key)
	data.Name = types.StringValue(env.Name)
	data.WithAlerts = types.BoolValue(env.WithAlerts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.UpdateEnvironment(ctx, &cronitor.Environment{
		Key:        plan.Key.ValueString(),
		Name:       plan.Name.ValueString(),
		WithAlerts: plan.WithAlerts.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update environment", err.Error())
		return
	}

	plan.Key = types.StringValue(env.Key)
	plan.Name = types.StringValue(env.Name)
	plan.WithAlerts = types.BoolValue(env.WithAlerts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEnvironment(ctx, data.Key.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete environment", err.Error())
		return
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
