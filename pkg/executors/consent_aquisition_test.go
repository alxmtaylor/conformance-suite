package executors

import (
	"fmt"
	"testing"

	"github.com/OpenBankingUK/conformance-suite/pkg/authentication"
	"github.com/OpenBankingUK/conformance-suite/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildParameters(t *testing.T) {

	permissions := map[string][]string{
		"tok0001": {"readbasic", "writebasic", "updatebasic"},
		"tok0002": {"updatebasic"},
		"tok0003": {},
	}

	buildstr := buildPermissionString(permissions["tok0001"])
	fmt.Println(buildstr)
	assert.Equal(t, `"readbasic","writebasic","updatebasic"`, buildstr)

	buildstr = buildPermissionString(permissions["tok0002"])
	fmt.Println(buildstr)
	assert.Equal(t, `"updatebasic"`, buildstr)

	buildstr = buildPermissionString(permissions["tok0003"])
	assert.Equal(t, ``, buildstr)
	fmt.Println(buildstr)

}

func TestConfigureHeadlessTokenEndpointAuth(t *testing.T) {
	tests := []struct {
		name       string
		grantType  string
		authMethod string
		wantBasic  bool
		wantError  bool
	}{
		{
			name:       "client credentials with tls client auth",
			grantType:  "client_credentials",
			authMethod: authentication.TlsClientAuth,
		},
		{
			name:       "authorization code with tls client auth",
			grantType:  authentication.GrantTypeAuthorizationCode,
			authMethod: authentication.TlsClientAuth,
		},
		{
			name:       "client credentials with client secret basic",
			grantType:  "client_credentials",
			authMethod: authentication.ClientSecretBasic,
			wantBasic:  true,
		},
		{
			name:       "unsupported authentication method",
			grantType:  "client_credentials",
			authMethod: "unsupported",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCase := model.MakeTestCase()
			testCase.Input.Method = "POST"
			testCase.Input.Endpoint = "$token_endpoint"
			testCase.Input.Headers["authorization"] = "Basic encoded-credentials"
			testCase.Input.FormData["grant_type"] = tt.grantType

			ctx := model.Context{
				"client_id":                  "client-id",
				"token_endpoint":             "https://example.test/token",
				"token_endpoint_auth_method": tt.authMethod,
			}

			err := configureHeadlessTokenEndpointAuth(&testCase, &ctx)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			req, err := testCase.Prepare(&ctx)
			require.NoError(t, err)
			if tt.wantBasic {
				assert.Equal(t, "Basic encoded-credentials", req.Header.Get("Authorization"))
			} else {
				assert.Empty(t, req.Header.Get("Authorization"))
				assert.Equal(t, "client-id", testCase.Input.FormData["client_id"])
			}
		})
	}
}

func TestHeadlessTokenProviderV4UsesVersionedConsentEndpoint(t *testing.T) {
	component, err := model.LoadComponent("headlessTokenProviderComponentV4.json")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(component.Tests), 3)

	assert.Equal(t, "/open-banking/$api-version/aisp/account-access-consents", component.Tests[1].Input.Endpoint)
	assert.Equal(t, "AWAU", component.Tests[1].Expect.Matches[0].Value)
	assert.Equal(t, "$responseType", component.Tests[2].Input.Claims["responseType"])
	assert.Equal(t, "$result_token", component.Tests[2].Input.Claims["state"])
	require.NotEmpty(t, component.Tests[2].Expect.ContextPut.Matches)
	assert.Equal(t, "code=([^&]*)", component.Tests[2].Expect.ContextPut.Matches[0].Regex)
}
