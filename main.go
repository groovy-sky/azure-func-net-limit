package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/urfave/cli/v2"
)

type accessList struct {
	ips   []string
	vnets []string
}

type resourceProperties struct {
	networkAcls *string `json:"networkAcls,omitempty"`
}

func azureLogin() (cred *azidentity.EnvironmentCredential, err error) {

	cred, err = azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		fmt.Println("[TODO] Azure login failure handeling")
	}
	return cred, err
}

func parseResourceId(resourceId string) (subscriptionId, resourceGroup, resourceProvider, resourceName string) {
	parts := strings.Split(resourceId, "/")
	subscriptionId = parts[2]
	resourceGroup = parts[4]
	resourceProvider = strings.Join(parts[6:8], "/")
	resourceName = parts[8]
	return subscriptionId, resourceGroup, resourceProvider, resourceName
}

func updateNetAcl(cred azcore.TokenCredential, resourceId, allowIPList, allowVNetList string) {

	subscriptionID, resourceGroupName, resourceProvider, resourceName := parseResourceId(resourceId)

	switch resourceProvider {
	case "Microsoft.Storage/storageAccounts":
		ctx := context.Background()

		storageAccountsClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)

		if err != nil {
			panic(err)
		}

		resource, err := storageAccountsClient.GetProperties(ctx, resourceGroupName, resourceName, &armstorage.AccountsClientGetPropertiesOptions{Expand: nil})
		if err != nil {
			panic(err)
		}

		for _, ip := range strings.Split(allowIPList, ";") {
			var exists bool
			for _, ipRule := range resource.Properties.NetworkRuleSet.IPRules {
				if *ipRule.IPAddressOrRange == ip {
					exists = true
					break
				}

			}
			if !exists {
				newRuleSet := &armstorage.IPRule{
					Action:           &[]string{"Allow"}[0],
					IPAddressOrRange: &[]string{ip}[0],
				}
				resource.Properties.NetworkRuleSet.IPRules = append(resource.Properties.NetworkRuleSet.IPRules, newRuleSet)
			}
		}

		for _, ipRule := range resource.Properties.NetworkRuleSet.IPRules {
			fmt.Println(*ipRule)
		}

		for _, v := range resource.Properties.NetworkRuleSet.IPRules {
			fmt.Println(*v.IPAddressOrRange, *v.Action)
		}

		resource.Properties.NetworkRuleSet.DefaultAction = &[]armstorage.DefaultAction{armstorage.DefaultActionDeny}[0]

		response, err := storageAccountsClient.Update(ctx, resourceGroupName, resourceName, armstorage.AccountUpdateParameters{Properties: &armstorage.AccountPropertiesUpdateParameters{NetworkRuleSet: resource.Properties.NetworkRuleSet}}, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(response)
	}

}

func getInputParams() (resList, ipList, vnetList string) {
	// returns input data from CLI
	app := &cli.App{
		Name:                 "aznet",
		Usage:                "CLI tool to set Azure PaaS network access",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "source",
				Aliases:     []string{"s"},
				Value:       "",
				Usage:       "Sources of PaaS resources list",
				Destination: &resList,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "ips",
				Aliases:     []string{"i"},
				Value:       "",
				Usage:       "Allowed IPs list",
				Destination: &ipList,
			},
			&cli.StringFlag{
				Name:        "vnets",
				Aliases:     []string{"v"},
				Value:       "",
				Usage:       "Allowed VNets list",
				Destination: &vnetList,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if resList == "" || (ipList == "" && vnetList == "") {
		log.Fatal("[ERR] Not enough imput parameters")
	}
	return resList, ipList, vnetList

}

func main() {

	res, ips, vnets := getInputParams()

	login, err := azureLogin()
	if err != nil {
		panic(err)
	}
	updateNetAcl(login, res, ips, vnets)

}
