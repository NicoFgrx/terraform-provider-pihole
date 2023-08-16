package provider

import (
	"context"
	"fmt"
	"time"

	pihole "github.com/NicoFgrx/pihole-api-go/api"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &CnameResource{}
	_ resource.ResourceWithConfigure   = &CnameResource{}
	_ resource.ResourceWithImportState = &CnameResource{}
)

// NewcnamerecordResource is a helper function to simplify the provider implementation.
func NewCnameResource() resource.Resource {
	return &CnameResource{}
}

// cnamerecordResource is the resource implementation.
type CnameResource struct {
	client *pihole.Client
}

// cnamerecordResourceModel maps the resource schema data.
type CnameResourceModel struct {
	// ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Domain      types.String `tfsdk:"domain"`
	Target      types.String `tfsdk:"target"`
}

// Metadata returns the resource type name.
func (r *CnameResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	fmt.Println(req.ProviderTypeName)
	resp.TypeName = req.ProviderTypeName + "_cname"
}

// Schema defines the schema for the resource.
func (r *CnameResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "CNAME Record resource for pihole",
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the cname record.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "Alias to use on CNAME",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:    true,
				Description: "Local managed DNS record",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CnameResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pihole.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *pihole.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create a new resource.
func (r *CnameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan CnameResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	data := pihole.CNAMERecordParams{
		Domain: plan.Domain.ValueString(),
		Target: plan.Target.ValueString(),
	}

	ctx = tflog.SetField(ctx, "api", r.client.APIKey)
	ctx = tflog.SetField(ctx, "url", r.client.BaseURL)
	ctx = tflog.SetField(ctx, "domain", data.Domain)
	ctx = tflog.SetField(ctx, "target", data.Target)

	// Create new cname record
	err := r.client.AddCustomCNAME(&data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating customcname",
			"Could not create customcname, unexpected error: "+err.Error(),
		)
	}

	// Map response body to schema and populate Computed attribute values

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read resource information.
func (r *CnameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state CnameResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refresh cname value
	cnamerecord, err := r.client.GetCustomCNAME(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pihole cnameRecord",
			"Could not read Pihole cnameRecord ID "+state.Domain.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Domain = types.StringValue(cnamerecord.Domain)
	state.Target = types.StringValue(cnamerecord.Target)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *CnameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update function will never be triggered because of Pihole API limitation
	return
}

func (r *CnameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state CnameResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Rebuild the cnameRecordParams to Delete
	to_delete := pihole.CNAMERecordParams{
		Domain: state.Domain.ValueString(),
		Target: state.Target.ValueString(),
	}

	// Delete existing record
	err := r.client.DeleteCustomCNAME(&to_delete)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting cname Record ",
			"Could not delete cname record, unexpected error: "+err.Error(),
		)
		return
	}

}

func (r *CnameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
