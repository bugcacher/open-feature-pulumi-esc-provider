package pulumi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	esc "github.com/pulumi/esc-sdk/sdk/go"
	"github.com/stretchr/testify/assert"
)

const (
	PROJECT_NAME = "of-pulumi-esc-provider-test"
	ENV_NAME     = "of-pulumi-esc-provider-test-env"
)

const (
	STRING_FLAG_KEY       = "SOME_STRING_FLAG"
	BOOL_FLAG_KEY         = "SOME_BOOL_FLAG"
	INT_FLAG_KEY          = "SOME_INT_FLAG"
	FLOAT_FLAG_KEY        = "SOME_FLOAT_FLAG"
	NON_EXISTING_FLAG_KEY = "NON_EXISTING_FLAG"

	STRING_FLAG_VALUE = "string-flag-value"
	BOOL_FLAG_VALUE   = true
	INT_FLAG_VALUE    = int64(50)
	FLOAT_FLAG_VALUE  = float64(0.5)

	DEFAULT_STRING_FLAG_VALUE = "default-value"
	DEFAULT_BOOL_FLAG_VALUE   = true
	DEFAULT_INT_FLAG_VALUE    = int64(10)
	DEFAULT_FLOAT_FLAG_VALUE  = float64(0.1)
)

var (
	provider *PulumiESCProvider
)

func TestPulumiESCProvider_Metadata(t *testing.T) {
	tests := []struct {
		name string
		p    *PulumiESCProvider
		want openfeature.Metadata
	}{
		{
			name: "provider-metadata",
			p:    provider,
			want: openfeature.Metadata{
				Name: ProviderName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Metadata(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PulumiESCProvider.Metadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPulumiESCProvider_Hooks(t *testing.T) {
	tests := []struct {
		name string
		p    *PulumiESCProvider
		want []openfeature.Hook
	}{
		{
			name: "provider-hooks-success",
			p:    provider,
			want: []openfeature.Hook{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Hooks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PulumiESCProvider.Hooks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPulumiESCProvider_Status(t *testing.T) {
	tests := []struct {
		name string
		p    *PulumiESCProvider
		want openfeature.State
	}{
		{
			name: "ready-state",
			p:    provider,
			want: openfeature.ReadyState,
		},
		{
			name: "error-state",
			p:    &PulumiESCProvider{state: openfeature.ErrorState},
			want: openfeature.ErrorState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Status(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PulumiESCProvider.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPulumiESCProvider_BooleanEvaluation(t *testing.T) {
	type args struct {
		ctx          context.Context
		flag         string
		defaultValue bool
		evalCtx      openfeature.FlattenedContext
	}
	tests := []struct {
		name string
		p    *PulumiESCProvider
		args args
		want openfeature.BoolResolutionDetail
	}{
		{
			name: "bool-flag-success",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         BOOL_FLAG_KEY,
				defaultValue: DEFAULT_BOOL_FLAG_VALUE,
			},
			want: openfeature.BoolResolutionDetail{
				Value: BOOL_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason: openfeature.StaticReason,
				},
			},
		},
		{
			name: "bool-flag-type-mismatch",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         INT_FLAG_KEY,
				defaultValue: DEFAULT_BOOL_FLAG_VALUE,
			},
			want: openfeature.BoolResolutionDetail{
				Value: DEFAULT_BOOL_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewTypeMismatchResolutionError(""),
				},
			},
		},
		{
			name: "bool-flag-missing",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         NON_EXISTING_FLAG_KEY,
				defaultValue: DEFAULT_BOOL_FLAG_VALUE,
			},
			want: openfeature.BoolResolutionDetail{
				Value: DEFAULT_BOOL_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewFlagNotFoundResolutionError(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.BooleanEvaluation(tt.args.ctx, tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Reason, got.ProviderResolutionDetail.Reason)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Variant, got.ProviderResolutionDetail.Variant)
			assert.Equal(t, tt.want.ProviderResolutionDetail.ResolutionDetail().ErrorCode, got.ProviderResolutionDetail.ResolutionDetail().ErrorCode)
		})
	}
}

func TestPulumiESCProvider_StringEvaluation(t *testing.T) {
	type args struct {
		ctx          context.Context
		flag         string
		defaultValue string
		evalCtx      openfeature.FlattenedContext
	}
	tests := []struct {
		name string
		p    *PulumiESCProvider
		args args
		want openfeature.StringResolutionDetail
	}{
		{
			name: "string-flag-success",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         STRING_FLAG_KEY,
				defaultValue: DEFAULT_STRING_FLAG_VALUE,
			},
			want: openfeature.StringResolutionDetail{
				Value: STRING_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason: openfeature.StaticReason,
				},
			},
		},
		{
			name: "string-flag-type-mismatch",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         INT_FLAG_KEY,
				defaultValue: DEFAULT_STRING_FLAG_VALUE,
			},
			want: openfeature.StringResolutionDetail{
				Value: DEFAULT_STRING_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewTypeMismatchResolutionError(""),
				},
			},
		},
		{
			name: "string-flag-missing",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         NON_EXISTING_FLAG_KEY,
				defaultValue: DEFAULT_STRING_FLAG_VALUE,
			},
			want: openfeature.StringResolutionDetail{
				Value: DEFAULT_STRING_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewFlagNotFoundResolutionError(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.StringEvaluation(tt.args.ctx, tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Reason, got.ProviderResolutionDetail.Reason)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Variant, got.ProviderResolutionDetail.Variant)
			assert.Equal(t, tt.want.ProviderResolutionDetail.ResolutionDetail().ErrorCode, got.ProviderResolutionDetail.ResolutionDetail().ErrorCode)
		})
	}
}

func TestPulumiESCProvider_FloatEvaluation(t *testing.T) {
	type args struct {
		ctx          context.Context
		flag         string
		defaultValue float64
		evalCtx      openfeature.FlattenedContext
	}
	tests := []struct {
		name string
		p    *PulumiESCProvider
		args args
		want openfeature.FloatResolutionDetail
	}{
		{
			name: "float-flag-success",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         FLOAT_FLAG_KEY,
				defaultValue: DEFAULT_FLOAT_FLAG_VALUE,
			},
			want: openfeature.FloatResolutionDetail{
				Value: FLOAT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason: openfeature.StaticReason,
				},
			},
		},
		{
			name: "float-flag-type-mismatch",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         BOOL_FLAG_KEY,
				defaultValue: DEFAULT_FLOAT_FLAG_VALUE,
			},
			want: openfeature.FloatResolutionDetail{
				Value: DEFAULT_FLOAT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewTypeMismatchResolutionError(""),
				},
			},
		},
		{
			name: "float-flag-missing",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         NON_EXISTING_FLAG_KEY,
				defaultValue: DEFAULT_FLOAT_FLAG_VALUE,
			},
			want: openfeature.FloatResolutionDetail{
				Value: DEFAULT_FLOAT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewFlagNotFoundResolutionError(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.FloatEvaluation(tt.args.ctx, tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Reason, got.ProviderResolutionDetail.Reason)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Variant, got.ProviderResolutionDetail.Variant)
			assert.Equal(t, tt.want.ProviderResolutionDetail.ResolutionDetail().ErrorCode, got.ProviderResolutionDetail.ResolutionDetail().ErrorCode)
		})
	}
}

func TestPulumiESCProvider_IntEvaluation(t *testing.T) {
	type args struct {
		ctx          context.Context
		flag         string
		defaultValue int64
		evalCtx      openfeature.FlattenedContext
	}
	tests := []struct {
		name string
		p    *PulumiESCProvider
		args args
		want openfeature.IntResolutionDetail
	}{
		{
			name: "int-flag-success",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         INT_FLAG_KEY,
				defaultValue: DEFAULT_INT_FLAG_VALUE,
			},
			want: openfeature.IntResolutionDetail{
				Value: INT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason: openfeature.StaticReason,
				},
			},
		},
		{
			name: "int-flag-type-mismatch",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         BOOL_FLAG_KEY,
				defaultValue: DEFAULT_INT_FLAG_VALUE,
			},
			want: openfeature.IntResolutionDetail{
				Value: DEFAULT_INT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewTypeMismatchResolutionError(""),
				},
			},
		},
		{
			name: "int-flag-missing",
			p:    provider,
			args: args{
				ctx:          context.TODO(),
				flag:         NON_EXISTING_FLAG_KEY,
				defaultValue: DEFAULT_INT_FLAG_VALUE,
			},
			want: openfeature.IntResolutionDetail{
				Value: DEFAULT_INT_FLAG_VALUE,
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewFlagNotFoundResolutionError(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IntEvaluation(tt.args.ctx, tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Reason, got.ProviderResolutionDetail.Reason)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Variant, got.ProviderResolutionDetail.Variant)
			assert.Equal(t, tt.want.ProviderResolutionDetail.ResolutionDetail().ErrorCode, got.ProviderResolutionDetail.ResolutionDetail().ErrorCode)
		})
	}
}

func TestPulumiESCProvider_ObjectEvaluation(t *testing.T) {
	type args struct {
		ctx          context.Context
		flag         string
		defaultValue interface{}
		evalCtx      openfeature.FlattenedContext
	}
	tests := []struct {
		name string
		p    *PulumiESCProvider
		args args
		want openfeature.InterfaceResolutionDetail
	}{
		{
			name: "object-flag-unimplemented",
			p:    provider,
			args: args{
				flag: "SOME_OBJECT_FLAG",
			},
			want: openfeature.InterfaceResolutionDetail{
				ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
					Reason:          openfeature.ErrorReason,
					ResolutionError: openfeature.NewGeneralResolutionError("ObjectEvaluation not implemented"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.ObjectEvaluation(tt.args.ctx, tt.args.flag, tt.args.defaultValue, tt.args.evalCtx)
			assert.Equal(t, tt.want.Value, got.Value)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Reason, got.ProviderResolutionDetail.Reason)
			assert.Equal(t, tt.want.ProviderResolutionDetail.Variant, got.ProviderResolutionDetail.Variant)
			assert.Equal(t, tt.want.ProviderResolutionDetail.ResolutionDetail().ErrorCode, got.ProviderResolutionDetail.ResolutionDetail().ErrorCode)
		})
	}
}

func TestMain(t *testing.M) {
	if err := setupTestProvider(); err != nil {
		fmt.Printf("Error during esc test provider setup: %v", err)
		os.Exit(1)
	}
	code := t.Run()
	if err := cleanup(); err != nil {
		fmt.Printf("Error during esc test provider cleanup: %v", err)
	}
	os.Exit(code)
}

// setupTestProvider requires the PULUMI_ORG and PULUMI_ACCESS_KEY environment variables to be set.
// If either of these variables is missing, the provider setup will fail.
func setupTestProvider() error {
	orgName := os.Getenv("PULUMI_ORG")
	if orgName == "" {
		return errors.New("PULUMI_ORG env variable can not be empty")
	}
	accessKey := os.Getenv("PULUMI_ACCESS_KEY")
	if accessKey == "" {
		return errors.New("PULUMI_ACCESS_KEY env variable can not be empty")
	}

	// Create or update test environment in Pulumi
	if err := createOrUpdatePulumiTestEnv(
		orgName,
		PROJECT_NAME,
		ENV_NAME,
		accessKey,
	); err != nil {
		return fmt.Errorf("failed to create/update pulumi test environment: %w", err)
	}

	customUrl, err := url.Parse("https://api.pulumi.com")
	if err != nil {
		return err
	}
	// Set test provider
	escProvider, err := NewPulumiESCProvider(
		orgName,
		PROJECT_NAME,
		ENV_NAME,
		accessKey,
		WithCustomBackendUrl(*customUrl),
	)
	if err != nil {
		return err
	}
	provider = escProvider
	return nil
}

func cleanup() error {
	orgName := os.Getenv("PULUMI_ORG")
	if orgName != "" {
		if err := removePulumiTestEnv(orgName, PROJECT_NAME, ENV_NAME); err != nil {
			return fmt.Errorf("failed to delete pulumi test environment: %w\n", err)
		}
	}
	return nil
}

func createOrUpdatePulumiTestEnv(orgName, projectName, envName, accessKey string) error {
	conf := esc.NewConfiguration()
	escClient := esc.NewClient(conf)
	escAuthCtx := esc.NewAuthContext(accessKey)

	err := escClient.CreateEnvironment(escAuthCtx, orgName, projectName, envName)
	// In case test environement already exists, we can skip the error and update the enviroment values
	if err != nil && !environemntAlreadyExists(err) {
		return err
	}

	_, err = escClient.UpdateEnvironment(escAuthCtx, orgName, projectName, envName, getTestEnvDefinition())
	if err != nil {
		return err
	}
	return nil
}

func getTestEnvDefinition() *esc.EnvironmentDefinition {
	return &esc.EnvironmentDefinition{
		Values: &esc.EnvironmentDefinitionValues{
			AdditionalProperties: map[string]interface{}{
				STRING_FLAG_KEY: STRING_FLAG_VALUE,
				BOOL_FLAG_KEY:   BOOL_FLAG_VALUE,
				INT_FLAG_KEY:    INT_FLAG_VALUE,
				FLOAT_FLAG_KEY:  FLOAT_FLAG_VALUE,
			},
		},
	}
}

func environemntAlreadyExists(err error) bool {
	var genErr *esc.GenericOpenAPIError
	if errors.As(err, &genErr) {
		type OpenAPIErrResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		var errResp OpenAPIErrResp
		err := json.Unmarshal(genErr.Body(), &errResp)
		if err != nil {
			return false
		}
		return errResp.Code == 409
	}
	return false
}

func removePulumiTestEnv(orgName, projectName, envName string) error {
	return provider.escClient.DeleteEnvironment(provider.escAuthCtx, orgName, projectName, envName)
}
