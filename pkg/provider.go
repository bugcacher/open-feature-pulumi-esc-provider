package pulumi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
	esc "github.com/pulumi/esc-sdk/sdk/go"
)

type FlagType string

const (
	FlagType_Bool    FlagType = "bool"
	FlagType_String  FlagType = "string"
	FlagType_Integer FlagType = "int64"
	FlagType_Float   FlagType = "float64"
	FlagType_Object  FlagType = "object"
)

const (
	ProviderName = "PulumiESCProvider"
)

// PulumiESCProvider implements the FeatureProvider interface and provides functions for evaluating flags
type PulumiESCProvider struct {
	state               openfeature.State
	orgName             string
	projectName         string
	envName             string
	escClient           *esc.EscClient
	escAuthCtx          context.Context
	escOpenEnvSessionId string
	customBackendUrl    *url.URL
}

type ProviderOption func(p *PulumiESCProvider)

func NewPulumiESCProvider(orgName, projectName, envName, accessKey string, opts ...ProviderOption) (*PulumiESCProvider, error) {
	provider := &PulumiESCProvider{
		state:       openfeature.NotReadyState,
		orgName:     orgName,
		projectName: projectName,
		envName:     envName,
	}
	for _, opt := range opts {
		opt(provider)
	}

	conf := esc.NewConfiguration()
	if provider.customBackendUrl != nil {
		customConf, err := esc.NewCustomBackendConfiguration(*provider.customBackendUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to initialise pulumi esc provider with custom backend url: %w", err)
		}
		conf = customConf
	}

	escClient := esc.NewClient(conf)
	escAuthCtx := esc.NewAuthContext(accessKey)
	env, err := escClient.OpenEnvironment(escAuthCtx, orgName, projectName, envName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise pulumi esc provider: %w", err)
	}

	provider.escClient = escClient
	provider.escAuthCtx = escAuthCtx
	provider.escOpenEnvSessionId = env.Id
	provider.state = openfeature.ReadyState
	return provider, nil
}

// WithCustomBackendUrl sets the specified URL as the Pulumi ESC backend API endpoint
func WithCustomBackendUrl(url url.URL) ProviderOption {
	return func(p *PulumiESCProvider) {
		p.customBackendUrl = &url
	}
}

// Metadata returns the metadata of the provider
func (p *PulumiESCProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: ProviderName,
	}
}

// Hooks returns a collection of openfeature.Hook defined by this provider
func (p *PulumiESCProvider) Hooks() []openfeature.Hook {
	return []openfeature.Hook{}
}

// Status expose the status of the provider
func (p *PulumiESCProvider) Status() openfeature.State {
	return p.state
}

// BooleanEvaluation returns a boolean flag
func (p *PulumiESCProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Bool)
	boolResolutionDetails := openfeature.BoolResolutionDetail{ProviderResolutionDetail: resolutionDetails}
	if value != nil {
		boolResolutionDetails.Value = value.(bool)
	} else {
		boolResolutionDetails.Value = defaultValue
	}
	return boolResolutionDetails
}

// StringEvaluation returns a string flag
func (p *PulumiESCProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_String)
	stringResolutionDetails := openfeature.StringResolutionDetail{ProviderResolutionDetail: resolutionDetails}
	if value != nil {
		stringResolutionDetails.Value = value.(string)
	} else {
		stringResolutionDetails.Value = defaultValue
	}
	return stringResolutionDetails
}

// FloatEvaluation returns a float flag
func (p *PulumiESCProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Float)
	floatResolutionDetails := openfeature.FloatResolutionDetail{ProviderResolutionDetail: resolutionDetails}
	if value != nil {
		floatResolutionDetails.Value = value.(float64)
	} else {
		floatResolutionDetails.Value = defaultValue
	}
	return floatResolutionDetails

}

// IntEvaluation returns an int flag
func (p *PulumiESCProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Integer)
	intResolutionDetails := openfeature.IntResolutionDetail{ProviderResolutionDetail: resolutionDetails}
	if value != nil {
		intResolutionDetails.Value = int64(value.(float64))
	} else {
		intResolutionDetails.Value = defaultValue
	}
	return intResolutionDetails

}

// ObjectEvaluation returns an object flag
func (p *PulumiESCProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: openfeature.NewGeneralResolutionError("ObjectEvaluation not implemented"),
		},
	}
}

// resolveValue retrieves a property value from the ESC service and validates its type.
// It returns the resolved value and resolution details, or an error if the property
// is not found, has a type mismatch, or any other error occurs.
func (p *PulumiESCProvider) resolveValue(ctx context.Context, propertyPath string, flagType FlagType) (interface{}, openfeature.ProviderResolutionDetail) {
	escValue, rawValue, err := p.escClient.ReadEnvironmentProperty(p.escAuthCtx, p.orgName, p.projectName, p.envName, p.escOpenEnvSessionId, propertyPath)
	if err != nil {
		var genErr *esc.GenericOpenAPIError
		if errors.As(err, &genErr) && isKeyNotFoundErr(genErr) {
			return nil, openfeature.ProviderResolutionDetail{
				Reason:          openfeature.ErrorReason,
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("%s not found", propertyPath)),
			}
		}
		return nil, openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: openfeature.NewGeneralResolutionError(err.Error()),
		}
	}
	if !validateType(rawValue, flagType) {
		return nil, openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: openfeature.NewTypeMismatchResolutionError(fmt.Sprintf("%s is of type %s, not of type %s", propertyPath, reflect.TypeOf(rawValue), flagType))}
	}
	return rawValue, openfeature.ProviderResolutionDetail{
		Reason: openfeature.StaticReason,
		FlagMetadata: openfeature.FlagMetadata{
			"secret": escValue.GetSecret(),
			"trace":  escValue.GetTrace(),
		},
	}
}

// validateType checks if the given raw value can be parsed into the given FlagType
func validateType(rawValue interface{}, flagType FlagType) bool {
	switch flagType {
	case FlagType_Bool:
		_, ok := rawValue.(bool)
		return ok
	case FlagType_String:
		_, ok := rawValue.(string)
		return ok
	case FlagType_Integer:
		// Integer values from Pulumi ESC are returned as float64 by the Pulumi go sdk
		_, ok := rawValue.(float64)
		return ok
	case FlagType_Float:
		_, ok := rawValue.(float64)
		return ok
	case FlagType_Object:
	}
	return false
}

// isKeyNotFoundErr determines whether the given GenericOpenAPIError indicates a 'key not found' condition.
func isKeyNotFoundErr(openApiErr *esc.GenericOpenAPIError) bool {
	type OpenAPIErrResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	var errResp OpenAPIErrResp
	err := json.Unmarshal(openApiErr.Body(), &errResp)
	if err != nil {
		return false
	}
	return errResp.Code == 400 && strings.Contains(errResp.Message, "not found")
}
