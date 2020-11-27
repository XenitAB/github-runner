# github-runner

## Information

Used to get a GitHub Self-hosted runner token using GitHub App.

The big reason for this is to be able to create a GitHub App that is limited to only self-hosted runners with no other access, being able to use this in automation projects and ephemeral buld agents.

The `stdout` will be the token that can be used to configure the self-hosted runner. It can either be configured through cli parameters or using secrets in Azure KeyVault (ENV and CLI authentication methods are supported).

Example usage can be found [here](https://github.com/XenitAB/packer-templates/tree/main/templates/azure/github-runner).

## Creating a GitHub App

The following is needed:

- App ID
- Installation ID
- Organization name
- Private Key

### Creating the app

Do the following:
GitHub -> Organization -> Settings -> Developer Settings -> GitHub Apps -> New GitHub App:

- GitHub App name: <unique name>
- Homepage URL: http://localhost
- Webhook > Active: [ ]
- Organization permissions > Self-hosted runners: Read & write
- Where can this GitHub App be installed? > [v] Only on this account
- Press Create GitHub App

Document the `App ID`.

### Installing the app in the organization

In the app, go to `Install App`:

- Select `Install` on organization
- Verify the permissions and organization, press `Install`

The URL you will be sent to will be something like: `https://github.com/organizations/<organization name>/settings/installations/<installation id>`

Document the `Installation ID` from the URL.

### Generating the private key

Go back to App Settings, in General go to Private Key and press Generate a private key.

Download the `Private key` and store it somewhere secure.

## Usage

### Without Azure KeyVault

```shell
go run cmd/github-runner/main.go --organization weaveworks-gitops-poc --app-id <id> --installation-id <id> --private-key-path <file>
```

### With Azure KeyVault

#### Configuring secret

```shell
az keyvault secret set --vault-name <Azure KeyVault Name> --name github-private-key --file <private key file>
az keyvault secret set --vault-name <Azure KeyVault Name> --name github-organization --value <GitHub organization name>
az keyvault secret set --vault-name <Azure KeyVault Name> --name github-app-id --file <GitHub Application ID>
az keyvault secret set --vault-name <Azure KeyVault Name> --name github-installation-id --file <GitHub Installation ID>
```

#### Azure ENV Authentication

```shell
go run cmd/github-runner/main.go --use-azure-keyvault --azure-keyvault-name <Azure KeyVault Name> --organization-kvsecret github-organization --app-id-kvsecret github-app-id --installation-id-kvsecret github-installation-id --private-key-kvsecret github-private-key
```

#### Azure CLI Authentication

```shell
go run cmd/github-runner/main.go --use-azure-keyvault --azure-keyvault-name <Azure KeyVault Name> --organization-kvsecret github-organization --app-id-kvsecret github-app-id --installation-id-kvsecret github-installation-id --private-key-kvsecret github-private-key --azure-auth CLI
```

### Other arguments

| Argument                     | Description                                                                          | Type / Options      | Default    | Required when              |
| ---------------------------- | ------------------------------------------------------------------------------------ | ------------------- | ---------- | -------------------------- |
| `--token-type`               | Token type to get from GitHub.                                                       | `REGISTER` `REMOVE` | `REGISTER` | Never                      |
| `--azure-auth`               | The Azure authentication method.                                                     | `ENV` `CLI`         | `CLI`      | Never                      |
| `--output`                   | How should the output be printed.                                                    | `TOKEN` `JSON`      | `TOKEN`    | Never                      |
| `--organization`             | Name of the GitHub organization.                                                     | `string`            | `""`       | `--useAzureKeyVault flase` |
| `--app-id`                   | Application ID of the GitHub App.                                                    | `string`            | `""`       | `--useAzureKeyVault flase` |
| `--installation-id`          | Installation ID of the GitHub App.                                                   | `int64`             | `0`        | `--useAzureKeyVault flase` |
| `--private-key-path`         | The private key (PEM format) from the GitHub App.                                    | `int64`             | `0`        | `--useAzureKeyVault flase` |
| `--useAzureKeyVault`         | Should parameters be extracted from Azure KeyVault.                                  | `bool`              | `false`    | Never                      |
| `--azure-keyvault-name`      | The name of the Azure KeyVault containing the secrets.                               | `string`            | `""`       | `--useAzureKeyVault true`  |
| `--organization-kvsecret`    | The key name of the Azure KeyVault secret containing the organization name value.    | `string`            | `""`       | `--useAzureKeyVault true`  |
| `--app-id-kvsecret`          | The name of the Azure KeyVault containing the secrets.                               | `string`            | `""`       | `--useAzureKeyVault true`  |
| `--installation-id-kvsecret` | The key name of the Azure KeyVault secret containing the Installation ID name value. | `string`            | `""`       | `--useAzureKeyVault true`  |
| `--private-key-kvsecret`     | he key name of the Azure KeyVault secret containing the GitHub Private Key value.    | `string`            | `""`       | `--useAzureKeyVault true`  |
| `--azure-keyvault-name`      | The name of the Azure KeyVault containing the secrets.                               | `string`            | `""`       | `--useAzureKeyVault true`  |
