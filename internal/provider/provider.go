package provider

import (
	"context"
	"os"

	pihole "github.com/NicoFgrx/pihole-api-go/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &piholeProvider{}
)

// piholeProviderModel maps provider schema data to a Go type.
type piholeProviderModel struct {
	Url   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &piholeProvider{
			version: version,
		}
	}
}

// piholeProvider is the provider implementation.
type piholeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *piholeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pihole"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *piholeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Pihole.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "URI for Pihole API. May also be provided via PIHOLE_API_URL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "Token for Pihole API. May also be provided via PIHOLE_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}
func (p *piholeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pihole client")

	// Retrieve provider data from configuration
	var config piholeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Url.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown PiHole API Host",
			"The provider cannot create the PiHole API client as there is an unknown configuration value for the PiHole API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PIHOLE_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown PiHole API Password",
			"The provider cannot create the PiHole API client as there is an unknown configuration value for the PiHole API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PIHOLE_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv("PIHOLE_API_URL")
	token := os.Getenv("PIHOLE_TOKEN")

	if !config.Url.IsNull() {
		url = config.Url.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing HashiCups API Host",
			"The provider cannot create the Pihole API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the PIHOLE_API_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing HashiCups API Password",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API password. "+
				"Set the password value in the configuration or use the PIHOLE_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new pihole client using the configuration values
	client := pihole.NewClient(url, token)

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *piholeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *piholeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsRecordResource,
	}
}
