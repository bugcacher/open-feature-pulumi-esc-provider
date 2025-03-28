package pulumi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-feature/go-sdk/openfeature"
	esc "github.com/pulumi/esc-sdk/sdk/go"
)

type FlagType int

const (
	FlagType_Bool FlagType = iota
	FlagType_String
	FlagType_Integer
	FlagType_Float
	FlagType_Object
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
}

type ProviderOption func(p *PulumiESCProvider)

func NewPulumiESCProvider(orgName, projectName, envName, accessKey string, opts ...ProviderOption) (*PulumiESCProvider, error) {
	conf := esc.NewConfiguration()
	escClient := esc.NewClient(conf)
	escAuthCtx := esc.NewAuthContext(accessKey)
	env, err := escClient.OpenEnvironment(escAuthCtx, orgName, projectName, envName)
	if err != nil {
		return nil, fmt.Errorf("failed to initiaze pulumi esc provider: %w", err)
	}
	provider := &PulumiESCProvider{
		state:               openfeature.ReadyState,
		orgName:             orgName,
		projectName:         projectName,
		envName:             envName,
		escClient:           escClient,
		escAuthCtx:          escAuthCtx,
		escOpenEnvSessionId: env.Id,
	}
	for _, opt := range opts {
		opt(provider)
	}
	return provider, nil
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
func (i PulumiESCProvider) Status() openfeature.State {
	return i.state
}

// BooleanEvaluation returns a boolean flag
func (p *PulumiESCProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Bool)
	return openfeature.BoolResolutionDetail{Value: value.(bool), ProviderResolutionDetail: resolutionDetails}
}

// StringEvaluation returns a string flag
func (p *PulumiESCProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_String)
	return openfeature.StringResolutionDetail{Value: value.(string), ProviderResolutionDetail: resolutionDetails}
}

// FloatEvaluation returns a float flag
func (p *PulumiESCProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Float)
	return openfeature.FloatResolutionDetail{Value: value.(float64), ProviderResolutionDetail: resolutionDetails}

}

// IntEvaluation returns an int flag
func (p *PulumiESCProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	value, resolutionDetails := p.resolveValue(ctx, flag, FlagType_Integer)
	return openfeature.IntResolutionDetail{Value: value.(int64), ProviderResolutionDetail: resolutionDetails}

}

// ObjectEvaluation returns an object flag
func (p *PulumiESCProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			ResolutionError: openfeature.NewGeneralResolutionError("ObjectEvaluation not implemented"),
		},
	}
}

func (p *PulumiESCProvider) resolveValue(ctx context.Context, propertyPath string, flagType FlagType) (interface{}, openfeature.ProviderResolutionDetail) {
	escValue, rawValue, err := p.escClient.ReadEnvironmentProperty(p.escAuthCtx, p.orgName, p.projectName, p.envName, p.escOpenEnvSessionId, propertyPath)
	if err != nil {
		var genErr *esc.GenericOpenAPIError
		if errors.As(err, &genErr) && isNotFoundErr(genErr) {
			return nil, openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewFlagNotFoundResolutionError(fmt.Sprintf("%s not found", propertyPath)),
			}
		}
		return nil, openfeature.ProviderResolutionDetail{ResolutionError: openfeature.NewGeneralResolutionError(err.Error())}
	}
	parsedValue, ok := typeChecker(rawValue, flagType)
	if !ok {
		return nil, openfeature.ProviderResolutionDetail{ResolutionError: openfeature.NewTypeMismatchResolutionError(fmt.Sprintf("%s not of type %s", propertyPath, flagType))}
	}
	return parsedValue, openfeature.ProviderResolutionDetail{
		Reason: openfeature.StaticReason,
		FlagMetadata: openfeature.FlagMetadata{
			"secret": escValue.GetSecret(),
			"trace":  escValue.GetTrace(),
		},
	}
}

func typeChecker(rawValue interface{}, flagType FlagType) (interface{}, bool) {
	switch flagType {
	case FlagType_Bool:
		parsedValue, ok := rawValue.(bool)
		return parsedValue, ok
	case FlagType_String:
		parsedValue, ok := rawValue.(string)
		return parsedValue, ok
	case FlagType_Integer:
		parsedValue, ok := rawValue.(int64)
		return parsedValue, ok
	case FlagType_Float:
		parsedValue, ok := rawValue.(float64)
		return parsedValue, ok
	case FlagType_Object:
	}
	return nil, false
}

func isNotFoundErr(err *esc.GenericOpenAPIError) bool {
	return strings.Contains(string(err.Body()), "not found")
}
