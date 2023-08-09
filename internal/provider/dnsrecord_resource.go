package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/NicoFgrx/pihole-api-go/api"
	pihole "github.com/NicoFgrx/pihole-api-go/api"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dnsrecordResource{}
	_ resource.ResourceWithConfigure   = &dnsrecordResource{}
	_ resource.ResourceWithImportState = &dnsrecordResource{}
)

// NewdnsrecordResource is a helper function to simplify the provider implementation.
func NewDnsRecordResource() resource.Resource {
	return &dnsrecordResource{}
}

// dnsrecordResource is the resource implementation.
type dnsrecordResource struct {
	client *pihole.Client
}

// dnsrecordResourceModel maps the resource schema data.
type dnsRecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Domain      types.String `tfsdk:"domain"`
	Ip          types.String `tfsdk:"ip"`
}

// Metadata returns the resource type name.
func (r *dnsrecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	fmt.Println(req.ProviderTypeName)
	resp.TypeName = req.ProviderTypeName + "_dnsrecord"
}

// Schema defines the schema for the resource.
func (r *dnsrecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "DNS Record resource for pihole",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Same as Domain attribute",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the dns record.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "FQDN of the Custom DNS Record",
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "IP address of the Custom DNS Record",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *dnsrecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *dnsrecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	data := api.DNSRecordParams{
		Domain: plan.Domain.ValueString(),
		IP:     plan.Ip.ValueString(),
	}

	ctx = tflog.SetField(ctx, "api", r.client.APIKey)
	ctx = tflog.SetField(ctx, "url", r.client.BaseURL)
	ctx = tflog.SetField(ctx, "domain", data.Domain)
	ctx = tflog.SetField(ctx, "ip", data.IP)

	// Create new dns record
	err := r.client.AddCustomDNS(&data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating customdns",
			"Could not create customdns, unexpected error: "+err.Error(),
		)
	}

	// Map response body to schema and populate Computed attribute values

	plan.ID = plan.Domain
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read resource information.
func (r *dnsrecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refresh dns value
	dnsrecord, err := r.client.GetCustomDNS(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pihole DNSRecord",
			"Could not read Pihole DNSRecord ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Domain = types.StringValue(dnsrecord.Domain)
	state.Ip = types.StringValue(dnsrecord.IP)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *dnsrecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state
	var state dnsRecordResourceModel
	diags_state := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags_state...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var plan dnsRecordResourceModel
	diags_plan := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags_plan...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	new := api.DNSRecordParams{
		Domain: plan.Domain.ValueString(),
		IP:     plan.Ip.ValueString(),
	}

	// Update existing record
	old := state.ID.ValueString()
	err := r.client.UpdateCustomDNS(old, &new)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating pihole dns record",
			"Could not update record, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated items from GetCustomDNS as DNSParms items are not
	dnsrecord, err := r.client.GetCustomDNS(new.Domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pihole CustomDNS",
			"Could not read Pihole DNSRecord ID "+new.Domain+": "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.Domain = types.StringValue(dnsrecord.Domain)
	plan.Ip = types.StringValue(dnsrecord.IP)

	plan.ID = plan.Domain
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *dnsrecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Rebuild the DNSRecordParams to Delete
	to_delete := pihole.DNSRecordParams{
		Domain: state.Domain.ValueString(),
		IP:     state.Ip.ValueString(),
	}

	// Delete existing record
	err := r.client.DeleteCustomDNS(&to_delete)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DNS Record ",
			"Could not delete dns record, unexpected error: "+err.Error(),
		)
		return
	}

}

func (r *dnsrecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
