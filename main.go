package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/aws/smithy-go/ptr"
)

// adjust the following variables
var (

	// az billing account list | jq '.[] | select(.displayName == "<YOUR-BILLING-ACCOUNT-DISPLAY-NAME>") | .name'
	billingAccountName = "<billingAccountName>"

	// az billing profile list --account-name <billingAccountName> | jq '.[] | select(.displayName == "<YOUR-BILLING-PROFILE-DISPLAY-NAME>") | .name'
	billingProfileName = "<billingProfileName>"

	// az billing invoice section list --account-name <billingAccountName> --profile-name <billingProfileName> | jq '.[] | select(.displayName == "<YOUR-INVOICE-SECTION-DISPLAY-NAME>") | .name'
	invoiceSectionName = "<invoiceSectionName>"

	// az account list | jq '.[] | select(.isDefault == true) | .id'
	subscriptionTenantID = "<subscriptionTenantID>"

	// owners email address
	subscriptionOwnerID = "<subscriptionOwnerID>"

	// the name of the subscription alias to create
	subscriptionAliasName = "<subscriptionAliasName>"
)

func main() {
	err := createSubscriptionAlias()
	if err != nil {
		panic(err)
	}
}

func connectionAzure() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func createSubscriptionAlias() error {
	creds, err := connectionAzure()
	if err != nil {
		return err
	}

	subAliasClient, err := armsubscription.NewAliasClient(creds, &policy.ClientOptions{})
	if err != nil {
		return err
	}

	workload := armsubscription.WorkloadProduction // important! does not work with WorkloadDevTest

	r, err := subAliasClient.BeginCreate(context.Background(), subscriptionAliasName, armsubscription.PutAliasRequest{
		Properties: &armsubscription.PutAliasRequestProperties{
			DisplayName: ptr.String(subscriptionAliasName),
			AdditionalProperties: &armsubscription.PutAliasRequestAdditionalProperties{
				SubscriptionTenantID: ptr.String(subscriptionTenantID),
				SubscriptionOwnerID:  ptr.String(subscriptionOwnerID),
			},
			BillingScope: ptr.String(fmt.Sprintf("/billingAccounts/%s/billingProfiles/%s/invoiceSections/%s", billingAccountName, billingProfileName, invoiceSectionName)),
			Workload:     &workload,
		},
	}, &armsubscription.AliasClientBeginCreateOptions{})
	if err != nil {
		return err
	}

	// wait for the operation to complete
	r.PollUntilDone(context.Background(), &runtime.PollUntilDoneOptions{
		Frequency: 5,
	})

	return nil
}
