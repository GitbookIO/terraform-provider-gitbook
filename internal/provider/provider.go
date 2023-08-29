package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const defaultIntegrationURL = "https://integrations.gitbook.com/v1/integrations/terraform/integration"

// gitBookProvider implements `provider.Provider`.
type gitBookProvider struct {
	// Version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// gitBookProviderModel describes the provider data model.
type gitBookProviderModel struct {
	APIBaseURL     types.String `tfsdk:"base_url"`
	IntegrationURL types.String `tfsdk:"integration_url"`
	AccessToken    types.String `tfsdk:"access_token"`
}

type integrationTokenEnvelope struct {
	// A short-lived JWT used to authenticate as a `terraform` installation.
	Token string `json:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &gitBookProvider{
			version: version,
		}
	}
}

func (p *gitBookProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gitbook"
	resp.Version = p.version
}

func (p *gitBookProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "GitBook API base URL (default: `https://api.gitbook.com`)",
				Optional:            true,
			},
			"integration_url": schema.StringAttribute{
				MarkdownDescription: "GitBook Terraform integration URL (default: `" + defaultIntegrationURL + "`)",
				Optional:            true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "GitBook Terraform integration access token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *gitBookProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	userAgent := "terraform-provider-gitbook/" + p.version

	var config gitBookProviderModel
	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIBaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown GitBook API base URL",
			"The provider cannot construct a GitBook API client as there is an unknown configuration value for the GitBook API base URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GITBOOK_API_BASE_URL environment variable.",
		)
	}

	if config.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown GitBook Terraform integration access token",
			"The provider cannot construct a GitBook API client as there is an unknown configuration value for the GitBook Terraform integration access token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GITBOOK_ACCESS_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables, but override with Terraform
	// configuration values if set.

	apiBaseURL := os.Getenv("GITBOOK_API_BASE_URL")
	accessToken := os.Getenv("GITBOOK_ACCESS_TOKEN")
	integrationURL := defaultIntegrationURL
	if env := os.Getenv("GITBOOK_INTEGRATION_URL"); env != "" {
		integrationURL = env
	}

	if !config.APIBaseURL.IsNull() {
		apiBaseURL = config.APIBaseURL.ValueString()
	}

	if !config.IntegrationURL.IsNull() {
		integrationURL = config.IntegrationURL.ValueString()
	}

	if !config.AccessToken.IsNull() {
		accessToken = config.AccessToken.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing GitBook Terraform integration access token",
			"The provider cannot construct a GitBook API client as there is a missing or empty value for the GitBook Terraform integration access token. "+
				"Set an `access_token` value in the configuration or use the GITBOOK_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Obtain a short-lived (installation scoped) GitBook API access token,
	// using the long-lived integration token that.
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodGet, integrationURL, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create HTTP request to obtain GitBook API access token",
			"An unexpected error occurred when constructing an HTTP request to obtain a short-lived GitBook API access token. "+
				"If the error is not clear, please contact GitBook support.\n\n"+
				"GitBook Terraform integration HTTP error: "+err.Error(),
		)
		return
	}

	tokenReq.Header.Set("Authorization", "Bearer "+accessToken)
	tokenReq.Header.Set("User-Agent", userAgent)
	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to obtain short-lived GitBook API access token",
			"An unexpected error occurred when obtaining a short-lived GitBook API access token. "+
				"If the error is not clear, please contact GitBook support.\n\n"+
				"GitBook Terraform integration HTTP error: "+err.Error(),
		)
		return
	}
	if tokenResp.StatusCode == http.StatusForbidden {
		resp.Diagnostics.AddError(
			"Invalid integration access token used",
			"The provided GitBook Terraform integration access token is invalid. "+
				"Visit the Terraform integration configuration on gitbook.com to obtain an access token.",
		)
		return
	} else if tokenResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to obtain short-lived GitBook API access token",
			"An unexpected HTTP response was received when obtaining a short-lived GitBook API access token. "+
				"If the error is not clear, please contact GitBook support.\n\n"+
				"GitBook Terraform integration HTTP response status: "+tokenResp.Status,
		)
		return
	}

	var tokenEnvelope integrationTokenEnvelope
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenEnvelope)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse token response from Terraform integration",
			"An unexpected error occurred parsing the HTTP response data from the Terraform integration. "+
				"If the error persists, please contact GitBook support.\n\n"+
				"Parsing error: "+err.Error(),
		)
		return

	}

	clientConfig := gitbook.NewConfiguration()
	// The API base URL can safely be an empty string; the underlying client
	// will fallback to the default `https://api.gitbook.com` in that case.
	if apiBaseURL != "" {
		hostVar := clientConfig.Servers[0].Variables["host"]
		hostVar.DefaultValue = apiBaseURL
		clientConfig.Servers[0].Variables["host"] = hostVar
	}

	// Using a custom default header simplifies usage of the client, as we don't
	// have to explicitly set a `gitbook.ContextAccessToken` context value for
	// each call.
	clientConfig.AddDefaultHeader("Authorization", "Bearer "+tokenEnvelope.Token)
	clientConfig.UserAgent = userAgent

	tflog.Debug(ctx, fmt.Sprintf("%+v", clientConfig))

	client := gitbook.NewAPIClient(clientConfig)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *gitBookProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEntityResource,
		NewEntitySchemaResource,
	}
}

func (p *gitBookProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewEntityDataSource,
		NewEntitySchemaDataSource,
	}
}
