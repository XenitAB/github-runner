package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	"github.com/bradleyfalzon/ghinstallation"
	github "github.com/google/go-github/v32/github"
	flag "github.com/spf13/pflag"
)

type authMethod string

const authMethodEnv authMethod = "ENV"
const authMethodCli authMethod = "CLI"

type config struct {
	organization                 string
	appID                        int64
	installationID               int64
	privateKeyPath               string
	useAzureKeyVault             bool
	azureKeyVaultName            string
	organizationKeyVaultSecret   string
	appIDKeyVaultSecret          string
	installationIDKeyVaultSecret string
	privateKeyKeyVaultSecret     string
	azureAuthenticationMethod    authMethod
}

func main() {
	runnerToken, err := getGitHubRunnerToken()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(runnerToken)
}

func getGitHubRunnerToken() (string, error) {
	config, err := getConfiguration()
	if err != nil {
		return "", err
	}

	transport := http.DefaultTransport
	var itr *ghinstallation.Transport

	if config.useAzureKeyVault {
		itr, config.organization, err = getGitHubTokenAzure(config, transport)
		if err != nil {
			return "", err
		}
	} else {
		itr, err = getGitHubToken(config, transport)
		if err != nil {
			return "", err
		}
	}

	client := github.NewClient(&http.Client{Transport: itr})

	ctx := context.Background()
	runnerToken, _, err := client.Actions.CreateOrganizationRegistrationToken(ctx, config.organization)
	if err != nil {
		return "", fmt.Errorf("ERROR: Unable to get registration token: %s", err)
	}

	return *runnerToken.Token, nil
}

func getConfiguration() (config, error) {
	_ = flag.Bool("debug", false, "Enables debug mode.")
	organization := flag.String("organization", "", "Name of the GitHub organization.")
	appID := flag.Int64("app-id", 0, "Application ID of the GitHub App.")
	installationID := flag.Int64("installation-id", 0, "Installation ID of the GitHub App.")
	privateKeyPath := flag.String("private-key-path", "", "The private key (PEM format) from the GitHub App.")
	useAzureKeyVault := flag.Bool("use-azure-keyvault", false, "Should parameters be extracted from Azure KeyVault.")
	azureKeyVaultName := flag.String("azure-keyvault-name", "", "The name of the Azure KeyVault containing the secrets.")
	organizationKeyVaultSecret := flag.String("organization-kvsecret", "", "The key name of the Azure KeyVault secret containing the organization name value.")
	appIDKeyVaultSecret := flag.String("app-id-kvsecret", "", "The key name of the Azure KeyVault secret containing the App ID name value.")
	installationIDKeyVaultSecret := flag.String("installation-id-kvsecret", "", "The key name of the Azure KeyVault secret containing the Installation ID name value.")
	privateKeyKeyVaultSecret := flag.String("private-key-kvsecret", "", "The key name of the Azure KeyVault secret containing the GitHub Private Key value.")
	azureAuthenticationMethod := flag.String("azure-auth", string(authMethodEnv), "The Azure authentication method.")
	flag.Parse()

	var err error

	authMethod, err := getAzureAuthMethod(*azureAuthenticationMethod)
	if err != nil {
		return config{}, fmt.Errorf("ERROR: Wrong auth method: %s\n", err)
	}

	return config{
		organization:                 *organization,
		appID:                        *appID,
		installationID:               *installationID,
		privateKeyPath:               *privateKeyPath,
		useAzureKeyVault:             *useAzureKeyVault,
		azureKeyVaultName:            *azureKeyVaultName,
		organizationKeyVaultSecret:   *organizationKeyVaultSecret,
		appIDKeyVaultSecret:          *appIDKeyVaultSecret,
		installationIDKeyVaultSecret: *installationIDKeyVaultSecret,
		privateKeyKeyVaultSecret:     *privateKeyKeyVaultSecret,
		azureAuthenticationMethod:    authMethod,
	}, nil
}

func getGitHubTokenAzure(config config, transport http.RoundTripper) (*ghinstallation.Transport, string, error) {
	var err error

	var authorizer autorest.Authorizer
	if config.azureAuthenticationMethod == authMethodEnv {
		authorizer, err = kvauth.NewAuthorizerFromEnvironment()
		if err != nil {
			return nil, "", fmt.Errorf("ERROR: unable to create vault authorizer: %v\n", err)
		}
	} else if config.azureAuthenticationMethod == authMethodCli {
		authorizer, err = kvauth.NewAuthorizerFromCLI()
		if err != nil {
			return nil, "", fmt.Errorf("ERROR: unable to create vault authorizer: %v\n", err)
		}
	} else {
		return nil, "", fmt.Errorf("ERROR: Valid values for --azure-auth is ENV and CLI. Received: %s", config.azureAuthenticationMethod)
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer

	orgnization, err := getKeyVaultSecret(basicClient, config.azureKeyVaultName, config.organizationKeyVaultSecret)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to get Organization secret from Azure KeyVault: %s", err)
	}

	appIDString, err := getKeyVaultSecret(basicClient, config.azureKeyVaultName, config.appIDKeyVaultSecret)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to get App ID secret from Azure KeyVault: %s", err)
	}

	appID, err := stringToInt64(appIDString)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to convert App ID from string to int64: %s", err)
	}

	installationIDString, err := getKeyVaultSecret(basicClient, config.azureKeyVaultName, config.installationIDKeyVaultSecret)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to get Installation ID secret from Azure KeyVault: %s", err)
	}

	installationID, err := stringToInt64(installationIDString)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to convert Installation ID from string to int64: %s", err)
	}

	privateKeyString, err := getKeyVaultSecret(basicClient, config.azureKeyVaultName, config.privateKeyKeyVaultSecret)
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to get Private Key secret from Azure KeyVault: %s", err)
	}

	itr, err := ghinstallation.New(transport, appID, installationID, []byte(privateKeyString))
	if err != nil {
		return nil, "", fmt.Errorf("ERROR: Unable to get token: %s", err)
	}

	return itr, orgnization, nil
}

func getGitHubToken(config config, transport http.RoundTripper) (*ghinstallation.Transport, error) {
	itr, err := ghinstallation.NewKeyFromFile(transport, config.appID, config.installationID, config.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to get token: %s", err)
	}

	return itr, nil
}

func getKeyVaultSecret(basicClient keyvault.BaseClient, vaultName, secname string) (string, error) {
	res, err := basicClient.GetSecret(context.Background(), "https://"+vaultName+".vault.azure.net", secname, "")
	if err != nil {
		return "", err
	}
	return *res.Value, nil
}

func stringToInt64(intString string) (int64, error) {
	n, err := strconv.ParseInt(intString, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func getAzureAuthMethod(authMethod string) (authMethod, error) {
	if authMethod == string(authMethodEnv) {
		return authMethodEnv, nil
	} else if authMethod == string(authMethodCli) {
		return authMethodCli, nil
	} else {
		return "", fmt.Errorf("ERROR: Valid values for --azure-auth is ENV and CLI. Received: %s", authMethod)
	}
}
