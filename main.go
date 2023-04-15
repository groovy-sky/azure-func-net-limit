package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
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

func getResource(cred azcore.TokenCredential) {

	ctx := context.Background()

	storageAccountsClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)

	if err != nil {
		panic(err)
	}

	resource, err := storageAccountsClient.GetProperties(ctx, resourceGroupName, resourceName, &armstorage.AccountsClientGetPropertiesOptions{Expand: nil})
	if err != nil {
		panic(err)
	}

	ipRule := &armstorage.IPRule{
		Action:           &[]string{"Allow"}[0],
		IPAddressOrRange: &[]string{"192.168.0.1/24"}[0],
	}

	resource.Properties.NetworkRuleSet.IPRules = append(resource.Properties.NetworkRuleSet.IPRules, ipRule)

	for _, v := range resource.Properties.NetworkRuleSet.IPRules {
		fmt.Println(*v.IPAddressOrRange, *v.Action)
	}
	fmt.Println(*resource.Properties.NetworkRuleSet.DefaultAction, *resource.Properties.NetworkRuleSet.Bypass, resource.Properties.NetworkRuleSet.IPRules, resource.Properties.NetworkRuleSet.ResourceAccessRules, resource.Properties.NetworkRuleSet.VirtualNetworkRules)
}

func main() {

	login, err := azureLogin()
	if err != nil {
		panic(err)
	}
	getResource(login)

}
