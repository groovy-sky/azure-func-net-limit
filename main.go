package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"io/ioutil"
	"net/http"

	"github.com/groovy-sky/net-limit-azure-paas/v2/netmerge"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/urfave/cli/v2"
)

// Login to Azure, using different kind of methods - credentials, managed identity
func azureLogin() (cred *azidentity.EnvironmentCredential, err error) {

	cred, err = azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		return cred, err
	}
	return cred, err
}

// Validates that input is valid for Azure IPv4/CIDR
func validateIPaddr(input string) (valid bool) {

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

// Removes duplicates from input slice
func unique(intSlice []string) (list []string) {
	keys := make(map[string]bool)
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Converts input string to slice of ips, with duplicate/whitespace removal
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

// Checks that input matches Azure Resource Id format
func validateResId(input string) bool {
	// Validates that input is valid Azure Resource Id using regex
	regex := `^\/subscriptions\/.{36}\/resourceGroups\/.*\/providers\/[a-zA-Z0-9]*.[a-zA-Z0-9]*\/[a-zA-Z0-9]*\/.*`
	r := regexp.MustCompile(regex)
	return r.MatchString(input)
}

// Gets Azure Resource's names of subscription, group, resource type etc.
func parseResourceId(resourceId string) (subscriptionId, resourceGroup, resourceProvider, resourceName string) {
	// takes Azure resource Id and parses sub id, group, resource type and name
	parts := strings.Split(resourceId, "/")
	subscriptionId = parts[2]
	resourceGroup = parts[4]
	resourceProvider = strings.Join(parts[6:8], "/")
	resourceName = parts[8]
	return subscriptionId, resourceGroup, resourceProvider, resourceName
}

func mergeIpLists(l1, l2 []string) []string {
	// Create a map to store the merged slice with no duplicates
	merged := make(map[string]bool)

	// Add elements from the first slice to the map
	for _, element := range l1 {
		merged[element] = true
	}

	// Add elements from the second slice to the map
	for _, element := range l2 {
		merged[element] = true
	}

	// Create a new slice to store the merged slice with no duplicates
	result := make([]string, 0, len(merged))

	// Append the unique elements to the new slice
	for element := range merged {
		result = append(result, element)
	}

	// Return the merged slice with no duplicates
	return result
}

// Whitelist specified IP list for PaaS resource, based on resource type
func SetPaasNet(cred azcore.TokenCredential, resourceId string, newIPList []string) (err error) {
	var newIpRuleSet []*armstorage.IPRule
	maxIpRules := 199 
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

		var oldIPList []string

		for _, ipRule := range resource.Properties.NetworkRuleSet.IPRules {
			oldIPList = append(oldIPList, *ipRule.IPAddressOrRange)
		}

		newIPList = mergeIpLists(newIPList, oldIPList)

		if len(newIPList) > len(oldIPList) {
			

			fmt.Println("Before: ", newIPList)
			for len(newIPList) > maxIpRules {
				newIPList, err = netmerge.MergeCIDRs(newIPList, uint8(maxIpRules))
				if err != nil {
					return err
				}
			}
			fmt.Println("After: ", newIPList)

			for _, ip := range newIPList {
				newRule := &armstorage.IPRule{
					Action:           &[]string{"Allow"}[0],
					IPAddressOrRange: &[]string{ip}[0],
				}
				newIpRuleSet = append(newIpRuleSet, []*armstorage.IPRule{newRule}...)

			}

			resource.Properties.NetworkRuleSet.IPRules = newIpRuleSet

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

func getIpsFromWeb(url string) (out string, err error) {  
	for _,link := range (strings.Split(url,";")){
		resp, err := http.Get(link)  
    if err != nil {  
        return "", err  
    }     
	defer resp.Body.Close()  
  
    body, err := ioutil.ReadAll(resp.Body)  
    if err != nil {  
        return "", err  
    } 
	out += string(body)

	}

    return out, err  
}  

	// returns input data from CLI
func getInputParams() (resList, ipList string) {
	var urlList string 
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
			},
			&cli.StringFlag{
				Name:        "url",
				Aliases:     []string{"u"},
				Value:       "",
				Usage:       "URL with allowed IPs",
				Destination: &urlList,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil || len(urlList+ipList) ==0  {
		log.Fatal("[ERR] Failed to get input:\n", err)
	}
	if strings.Contains(urlList,"https://") {
		urlList,err = getIpsFromWeb(urlList)
		if err != nil {
			log.Fatal("[ERR] Failed to download IP lists:\n", err)
		}
		ipList = ipList + ";" + urlList
	}
	fmt.Println(urlList)

	return resList, ipList

}

func main() {

	res, ips := getInputParams()

	login, err := azureLogin()
	if err != nil {
		log.Fatal("[ERR] Failed to login:\n", err)
	}
	fmt.Println(SetPaasNet(login, res, parseIPaddr(ips)))

}
