# Environment Setup Guide for Pulumi ESC

This guide walks you through setting up your Pulumi ESC environment for securely accessing configuration and secrets with OpenFeature in your Go application.

---

## 1. Create a Pulumi Organization or Project _(If not already created)_

Before setting up an environment, make sure you have a Pulumi organization and a project.

You can create both through the [Pulumi Cloud Console](https://app.pulumi.com/)

---

## 2. Create an ESC Environment

You can create environments via:

- The Pulumi Cloud Console

  <img width="1423" alt="Screenshot 2025-04-06 at 12 49 03 PM" src="https://github.com/user-attachments/assets/e10d245a-fa7e-4bb0-9b3c-f787eaa870d9" />


- Pulumi ESC CLI

  ```bash
  esc env init [<org-name>/][<project-name>/]<environment-name>
  ```

  [Managing Pulumi ESC Environments via CLI](https://www.pulumi.com/docs/esc/environments/working-with-environments/)

---

## 3. Add Static Configs and Secrets

You can define plaintext configuration values and encrypted secrets using the Pulumi Cloud console or Pulumi ESC CLI.

Here we are using Pulumi Cloud Console to define plaintext configs `USERS_DB_MONGO_URL`, `MAX_CONNECTIONS`, `DEBUG_MODE`, and `CPU_THRESHOLD` and encrypted secret `OPENAI_API_KEY` inside the root key `configs`.

```yaml
values:
  configs:
    USERS_DB_MONGO_URL: mongodb://db-host:27017/users
    MAX_CONNECTIONS: 100
    DEBUG_MODE: true
    CPU_THRESHOLD: 0.9
    OPENAI_API_KEY:
      fn::secret:
        ciphertext: <your-encrypted-secret>
```
<img width="1419" alt="Screenshot 2025-04-06 at 12 39 22 PM" src="https://github.com/user-attachments/assets/22e0dfa0-bf8d-4e75-afc3-23115e03603e" />

---

## 4. Set Up Cloud Provider Secrets (e.g., AWS)

To retrieve secrets from cloud providers like AWS, you can:

- Use [static access tokens](https://www.pulumi.com/docs/esc/integrations/dynamic-login-credentials/aws-login/#awsloginstatic)
- Or configure [OIDC authentication](https://www.pulumi.com/docs/esc/integrations/dynamic-login-credentials/aws-login/#awsloginstatic)

Here, we are using OIDC authentication to connect to AWS and fetch secrets from AWS Secrets Managger and parameters from AWS Parameters Store.

### YAML Configuration Example

```yaml
values:
  aws:
    login:
      fn::open::aws-login:
        oidc:
          roleArn: arn:aws:iam::<account-id>:role/pulumi-secrets-role
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
```

<img width="1421" alt="Screenshot 2025-04-06 at 12 52 16 PM" src="https://github.com/user-attachments/assets/6b10d226-767f-44c6-9e08-09f1e63fe0a4" />



Refer [AWS Secrets Provider Docs](https://www.pulumi.com/docs/esc/providers/aws-secrets/) and [AWS Parameter Store Docs](https://www.pulumi.com/docs/esc/providers/aws-parameter-store/) to learn more about importing secrets from AWS services.


To learn how to configure OpenID Connect (OIDC) between Pulumi Cloud and Cloud Vendors like AWS, GCP, Azure, etc, refer [Dynamic login credentials documentation](https://www.pulumi.com/docs/esc/integrations/dynamic-login-credentials/).



---

## 5. Generate and Use Pulumi Access Token

Pulumi offers the following types of Access Tokens:
- Personal Tokens
- Organizational Tokens
- Team Tokens


Here, we are generating a new personal access token from the Pulumi Dashboard under **Settings > Access Tokens**


Export the token for use in your application:

```bash
export PULUMI_ACCESS_TOKEN="pul-xxxxxxxxxxxxxxxx"
```

Refer [Managing Pulumi Access Tokens](https://www.pulumi.com/docs/pulumi-cloud/access-management/access-tokens/) to learn more about managing Access Tokens in Pulumi.

---

## Ready to Use With OpenFeature

You’ve now successfully set up your Pulumi ESC environment. You can now integrate these secrets and configs using the OpenFeature Pulumi ESC provider.

Refer to the [README.md](./README.md) for usage examples and implementation guidance.
