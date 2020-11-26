package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	flag.Parse()

	config := config{
		organization:   *organization,
		appID:          *appID,
		installationID: *installationID,
		privateKeyPath: *privateKeyPath,
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, config.appID, config.installationID, config.privateKeyPath)
	if err != nil {
		fmt.Printf("ERROR: Unable to get token: %s", err)
		os.Exit(1)
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
