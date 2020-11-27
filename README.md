# github-runner

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
