# OpenFeature Pulumi ESC Provider

<p>
  <img alt="OpenFeature + Pulumi ESC" src="https://img.shields.io/badge/openfeature-provider-blue" />
  <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/bugcacher/open-feature-pulumi-esc-provider" />
  <img alt="PkgGoDev" src="https://pkg.go.dev/badge/github.com/bugcacher/open-feature-pulumi-esc-provider" />
  <img alt="Coverage" src="https://img.shields.io/codecov/c/github/bugcacher/open-feature-pulumi-esc-provider" />
  <img alt="GitHub Release" src="https://img.shields.io/github/v/release/bugcacher/open-feature-pulumi-esc-provider" />
  <img alt="License" src="https://img.shields.io/github/license/bugcacher/open-feature-pulumi-esc-provider" />
  <img alt="GitHub Stars" src="https://img.shields.io/github/stars/your-org/your-repo?style=social" />
</p>

An [OpenFeature](https://openfeature.dev/docs/reference/intro#what-is-openfeature) provider written in Go for securely accessing **secrets and configuration values** from [Pulumi ESC](https://www.pulumi.com/esc/). This [provider](https://openfeature.dev/docs/reference/concepts/provider) allows your application to retrieve values like API keys, DB URLs, feature flags, and more — in a typed, structured, and cloud-agnostic way.

Using **Pulumi ESC** as the source of truth and **OpenFeature** as the standard interface, this tool simplifies environment configuration and makes it effortless to manage secrets across multiple cloud providers.

---

## Features

- Strongly-typed config access (`string`, `bool`, `int`, `float`)
- Built-in support for default fallback values
- Fetch secrets/configs from AWS, GCP, Azure or any other cloud vendor (via Pulumi ESC)
- Minimal setup using Pulumi ESC with OIDC authentication
- Fully compatible with the OpenFeature SDK in Go

---

## Installation

```bash
go get github.com/bugcacher/open-feature-pulumi-esc-provider
```

## Pulumi ESC Setup

You must configure a Pulumi ESC environment with the secrets and configuration values you want to expose. Here's an example configuration file:

```yaml
values:
  aws:
    login:
      fn::open::aws-login:
        oidc:
          roleArn: arn:aws:iam::8426XXXXX712:role/pulumi-read-role
          sessionName: pulumi-esc-session
    secrets:
      fn::open::aws-secrets:
        region: us-east-1
        login: ${aws.login}
        get:
          GITHUB_ACCESS_TOKEN:
            secretId: GITHUB_ACCESS_TOKEN
    params:
      fn::open::aws-parameter-store:
        region: us-east-1
        login: ${aws.login}
        get:
          GOOGLE_API_KEY:
            name: GOOGLE_API_KEY
  configs:
    USERS_DB_MONGO_URL: mongodb://123.0.0.789:27017/users_db
    MAX_CONNECTIONS: 101
    DEBUG_MODE: true
    CPU_THRESHOLD: 0.85
    OPENAI_API_KEY:
      fn::secret:
        ciphertext: ZXNjeAAAAAEAAAEA+en91vxgXnLsa/YKPF7JMoB6nccq0S6t2XboLIUcqRJOP25J+TCKvnwJiCcykh+x5FFWIoL8DmOWw78ZoHn5qPzCVw==
```

For a comprehensive setup on Pulumi ESC configuration, refer [pulumi-esc-guide.md](./pulumi-esc-guide.md)

## Usage Example

Here’s a minimal example of how to use this provider in a Go app:

```go

package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/bugcacher/open-feature-pulumi-esc-provider/pkg/pulumi"
)

func main() {
	// NOTE: Never hardcode sensitive values like access tokens.
	// Use environment variables or secret management tools instead.
	accessToken := "pul-xxxxxxxxxxxxxxxx" // Preferably from os.Getenv("PULUMI_ACCESS_TOKEN")

	// Pulumi ESC environment configuration
	orgName := "your-org-name"
	projectName := "your-project-name"
	envName := "your-env-name"

	// (Optional) Custom backend URL if you're self-hosting Pulumi ESC
	backendURL, _ := url.Parse("https://api.pulumi.com")

	// Initialize the OpenFeature provider backed by Pulumi ESC
	provider, err := pulumi.NewPulumiESCProvider(
		orgName,
		projectName,
		envName,
		accessToken,
		pulumi.WithCustomBackendUrl(*backendURL), // Optional ProviderOption
	)
	if err != nil {
		fmt.Printf("Failed to initialize Pulumi ESC provider: %v\n", err)
		return
	}

	// Register the provider globally
	if err := openfeature.SetProviderAndWait(provider); err != nil {
		fmt.Printf("Failed to set provider: %v\n", err)
		return
	}

	// Create a new OpenFeature client
	client := openfeature.NewClient("example-app")
	ctx := context.Background()


	// Fetch secret stored in AWS Secrets Manager via Pulumi ESC
	awsSecret, _ := client.StringValueDetails(ctx,
		"aws.secrets.GITHUB_ACCESS_TOKEN",
		"default-github-token", openfeature.EvaluationContext{},
	)
	fmt.Println("GitHub Access Token (AWS Secrets Manager):", awsSecret.Value)


	// Fetch parameter from AWS Parameter Store via Pulumi ESC
	awsParam, _ := client.StringValueDetails(ctx,
		"aws.params.GOOGLE_API_KEY",
		"default-google-api-key", openfeature.EvaluationContext{},
	)
	fmt.Println("Google API Key (AWS Parameter Store):", awsParam.Value)


	// Fetch string config value from Pulumi ESC
	dbURL, _ := client.StringValueDetails(ctx,
		"configs.USERS_DB_MONGO_URL",
		"mongodb://localhost:27017", openfeature.EvaluationContext{},
	)
	fmt.Println("Users DB URL (Pulumi ESC config value):", dbURL.Value)


	// Fetch integer config value from Pulumi ESC
	maxConnections, _ := client.IntValueDetails(ctx,
		"configs.MAX_CONNECTIONS",
		50, openfeature.EvaluationContext{},
	)
	fmt.Println("Max Connections (Pulumi ESC config value):", maxConnections.Value)


	// Fetch boolean config value from Pulumi ESC
	debugMode, _ := client.BooleanValueDetails(ctx,
		"configs.DEBUG_MODE",
		false, openfeature.EvaluationContext{},
	)
	fmt.Println("Debug Mode Enabled (Pulumi ESC config value):", debugMode.Value)


	// Fetch float config value from Pulumi ESC
	cpuThreshold, _ := client.FloatValueDetails(ctx,
		"configs.CPU_THRESHOLD",
		0.75, openfeature.EvaluationContext{},
	)
	fmt.Println("CPU Threshold (Pulumi ESC config value):", cpuThreshold.Value)


    // Fetch secret value encrypted within Pulumi ESC
	secretValue, _ := client.StringValueDetails(ctx,
		"configs.OPENAI_API_KEY",
		"sk-12345", openfeature.EvaluationContext{},
	)
	fmt.Println("Encrypted Secret (Pulumi ESC secret value):", secretValue.secretValue)
	fmt.Println("Is secret value? ", a.FlagMetadata.GetBool("secret"))
}

```

## Options

- **WithCustomBackendUrl**: It sets the specified URL as the Pulumi ESC backend API endpoint.

## Why Use This?

Environment variables and secrets are traditionally handled via .env files, CI/CD variables, or K8s secrets—each with its own limitations and risks.

With Pulumi ESC, you get:

- Centralized, secure config management
- Cross-cloud support (AWS/GCP/Azure)
  -First-class support for secret providers and typed configs
- A better developer experience using Pulumi AI and ESC SDKs
- And with OpenFeature, your application remains portable across config providers with a consistent API interface.

## Related Links

- [Pulumi ESC Documentation](https://www.pulumi.com/esc/)
- [Pulumi ESC Go SDK](https://github.com/pulumi/esc-sdk)
- [OpenFeature Docs](https://openfeature.dev/docs/reference/intro)
- [OpenFeature Go SDK](https://github.com/open-feature/go-sdk)

## Contributing

Contributions welcome! Feel free to:

- Open issues
- Submit PRs
- Star ⭐ the repo if it helped you
