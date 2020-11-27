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

type config struct {
	organization   string
	appID          int64
	installationID int64
	privateKeyPath string
}

func main() {
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
	azureAuthenticationMethod := flag.String("azure-auth", "ENV", "The Azure authentication method. Valid value: ENV | CLI")
	flag.Parse()

	config := config{}

	tr := http.DefaultTransport
	var itr *ghinstallation.Transport
	var err error

	if *useAzureKeyVault {
		var authorizer autorest.Authorizer
		if *azureAuthenticationMethod == "ENV" {
			authorizer, err = kvauth.NewAuthorizerFromEnvironment()
			if err != nil {
				fmt.Printf("ERROR: unable to create vault authorizer: %v\n", err)
				os.Exit(1)
			}
		} else if *azureAuthenticationMethod == "CLI" {
			authorizer, err = kvauth.NewAuthorizerFromCLI()
			if err != nil {
				fmt.Printf("ERROR: unable to create vault authorizer: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("ERROR: Valid values for --azure-auth is ENV and CLI. Received: %s", *azureAuthenticationMethod)
			os.Exit(1)
		}

		basicClient := keyvault.New()
		basicClient.Authorizer = authorizer

		config.organization, err = getKeyVaultSecret(basicClient, *azureKeyVaultName, *organizationKeyVaultSecret)
		if err != nil {
			fmt.Printf("ERROR: Unable to get Organization secret from Azure KeyVault: %s", err)
			os.Exit(1)
		}

		appIDString, err := getKeyVaultSecret(basicClient, *azureKeyVaultName, *appIDKeyVaultSecret)
		if err != nil {
			fmt.Printf("ERROR: Unable to get App ID secret from Azure KeyVault: %s", err)
			os.Exit(1)
		}

		config.appID, err = stringToInt64(appIDString)
		if err != nil {
			fmt.Printf("ERROR: Unable to convert App ID from string to int64: %s", err)
			os.Exit(1)
		}

		installationIDString, err := getKeyVaultSecret(basicClient, *azureKeyVaultName, *installationIDKeyVaultSecret)
		if err != nil {
			fmt.Printf("ERROR: Unable to get Installation ID secret from Azure KeyVault: %s", err)
			os.Exit(1)
		}

		config.installationID, err = stringToInt64(installationIDString)
		if err != nil {
			fmt.Printf("ERROR: Unable to convert Installation ID from string to int64: %s", err)
			os.Exit(1)
		}

		privateKeyString, err := getKeyVaultSecret(basicClient, *azureKeyVaultName, *privateKeyKeyVaultSecret)
		if err != nil {
			fmt.Printf("ERROR: Unable to get Private Key secret from Azure KeyVault: %s", err)
			os.Exit(1)
		}

		itr, err = ghinstallation.New(tr, config.appID, config.installationID, []byte(privateKeyString))
		if err != nil {
			fmt.Printf("ERROR: Unable to get token: %s", err)
			os.Exit(1)
		}
	} else {
		config.organization = *organization
		config.appID = *appID
		config.installationID = *installationID
		config.privateKeyPath = *privateKeyPath

		itr, err = ghinstallation.NewKeyFromFile(tr, config.appID, config.installationID, config.privateKeyPath)
		if err != nil {
			fmt.Printf("ERROR: Unable to get token: %s", err)
			os.Exit(1)
		}
	}

	client := github.NewClient(&http.Client{Transport: itr})

	ctx := context.Background()
	runnerToken, res, err := client.Actions.CreateOrganizationRegistrationToken(ctx, config.organization)
	if err != nil {
		fmt.Println(res)
		fmt.Printf("ERROR: Unable to get registration token: %s", err)
		os.Exit(1)
	}

	fmt.Println(*runnerToken.Token)
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
