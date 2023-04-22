package main

import (
	"azure-func-net-limit/netmerge"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/urfave/cli/v2"
)

func azureLogin() (cred *azidentity.EnvironmentCredential, err error) {
	// Tries to get Azure credentials
	cred, err = azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		fmt.Println("[TODO] Azure login failure handeling")
	}
	return cred, err
}

func validateIPaddr(input string) (valid bool) {
	// Validates that input is valid IPv4/CIDR

	// Checks that addr matches with IPv4 struct
	regex := `^(?:[0-9]{1,3}\.){3}[0-9]{1,3}.*`
	r := regexp.MustCompile(regex)

	if r.MatchString(input) {
		switch strings.Contains(input, "/") {
		case false:
			input = input + "/32"
			fallthrough
		default:
			ip, _, err := net.ParseCIDR(input)
			if err == nil && !ip.IsPrivate() {
				valid = true
			}

		}
	}

	return valid
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func parseIPaddr(ips string) (result []string) {
	// Parses input values, splits, validates if these are valid IPv4 and stores to arr

	// removing whitespaces
	ips = strings.ReplaceAll(ips, " ", "")
	// replacing newlines with semicolons
	ips = strings.ReplaceAll(ips, "\n", ";")
	for _, ip := range strings.Split(ips, ";") {
		if validateIPaddr(ip) {
			result = append(result, ip)
		}
	}
	result = unique(result)
	return result
}

func validateResId(input string) bool {
	// Validates that input is valid Azure Resource Id using regex
	regex := `^\/subscriptions\/.{36}\/resourceGroups\/.*\/providers\/[a-zA-Z0-9]*.[a-zA-Z0-9]*\/[a-zA-Z0-9]*\/.*`
	r := regexp.MustCompile(regex)
	return r.MatchString(input)
}

func parseResourceId(resourceId string) (subscriptionId, resourceGroup, resourceProvider, resourceName string) {
	// takes Azure resource Id and parses sub id, group, resource type and name
	parts := strings.Split(resourceId, "/")
	subscriptionId = parts[2]
	resourceGroup = parts[4]
	resourceProvider = strings.Join(parts[6:8], "/")
	resourceName = parts[8]
	return subscriptionId, resourceGroup, resourceProvider, resourceName
}

func updateNetAcl(cred azcore.TokenCredential, resourceId string, allowIPList []string) (err error) {
	var whitelistIps []string
	// Takes as input resource id and tries to apply to it IP/VNet restrictions

	if !validateResId(resourceId) {
		return (fmt.Errorf("[ERR]: %s is malformed", resourceId))
	}

	subscriptionID, resourceGroupName, resourceProvider, resourceName := parseResourceId(resourceId)

	switch resourceProvider {
	case "Microsoft.Storage/storageAccounts":
		ctx := context.Background()

		storageAccountsClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)

		if err != nil {
			return (fmt.Errorf("[ERR]: Couldn't access %s\n%e", resourceName, err))
		}

		resource, err := storageAccountsClient.GetProperties(ctx, resourceGroupName, resourceName, &armstorage.AccountsClientGetPropertiesOptions{Expand: nil})
		if err != nil {
			return (fmt.Errorf("[ERR]: Couldn't get properties of %s\n%e", resourceName, err))
		}

		oldIPRuleSetSize := len(resource.Properties.NetworkRuleSet.IPRules)

		// Appends allowed IPs w check for duplicates
		for _, ip := range allowIPList {
			var exists bool
			for _, ipRule := range resource.Properties.NetworkRuleSet.IPRules {
				if *ipRule.IPAddressOrRange == ip {
					exists = true
					break
				}

			}
			if !exists {
				whitelistIps = append(whitelistIps, ip)
			}
		}

		for len(resource.Properties.NetworkRuleSet.IPRules) > 10 {
			whitelistIps, err = netmerge.MergeCIDRs(whitelistIps, 10)
			if err != nil {
				return err
			}
		}

		if oldIPRuleSetSize < len(resource.Properties.NetworkRuleSet.IPRules) {

			resource.Properties.NetworkRuleSet.DefaultAction = &[]armstorage.DefaultAction{armstorage.DefaultActionDeny}[0]

			response, err := storageAccountsClient.Update(ctx, resourceGroupName, resourceName, armstorage.AccountUpdateParameters{Properties: &armstorage.AccountPropertiesUpdateParameters{NetworkRuleSet: resource.Properties.NetworkRuleSet}}, nil)
			if err != nil {
				return err
			}
			fmt.Println(response)
		}

	}
	return err

}

func getInputParams() (resList, ipList string) {
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
				Usage:       "PaaS resources list",
				Destination: &resList,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "ips",
				Aliases:     []string{"i"},
				Value:       "",
				Usage:       "Allowed IPs",
				Destination: &ipList,
				Required:    true,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("[ERR] Failed to get input:\n", err)
	}

	return resList, ipList

}

func main() {

	res, ips := getInputParams()

	login, err := azureLogin()
	if err != nil {
		log.Fatal("[ERR] Failed to login:\n", err)
	}
	fmt.Println(updateNetAcl(login, res, parseIPaddr(ips)))

}
