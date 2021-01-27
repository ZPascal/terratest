// +build azure

// NOTE: We use build tags to differentiate azure testing because we currently do not have azure access setup for
// CircleCI.

package test

import (
	"strings"
	"testing"
	"os"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAzureServiceBusExample(t *testing.T) {
	t.Parallel()

	uniquePostfix := strings.ToLower(random.UniqueId())
	
	// website::tag::1:: Configure Terraform setting up a path to Terraform code.
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../examples/azure/terraform-azure-servicebus-example",
		Vars: map[string]interface{}{
			"postfix": uniquePostfix,
		},
	}

	// website::tag::4:: At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// website::tag::2:: Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// website::tag::3:: Run `terraform output` to get the values of output variables
	expectedTopicSubscriptionsMap := terraform.OutputMapOfObjects(t, terraformOptions, "topics")
	expectedNamespaceName := terraform.Output(t, terraformOptions, "namespace_name")
	expectedResourceGroup := terraform.Output(t, terraformOptions, "resource_group")

	for topicName, topicsMap := range expectedTopicSubscriptionsMap {
		subscriptionNamesFromAzure, err := azure.ListTopicSubscriptionsNameE(
			os.Getenv("ARM_SUBSCRIPTION_ID"),
			expectedNamespaceName,
			expectedResourceGroup,
			topicName)
			
		if err != nil {
			t.Fatal(err)
		}

		subscriptionsMap := topicsMap.(map[string]interface{})["subscriptions"].(map[string]interface{})
		subscriptionNamesFromOutput := getMapKeylist(subscriptionsMap)
		// each subscription from the output should also exist in Azure
		assert.Equal(t, len(*subscriptionNamesFromOutput), len(*subscriptionNamesFromAzure))
		for _, subscrptionName := range *subscriptionNamesFromOutput {
			assert.Contains(t, *subscriptionNamesFromAzure, subscrptionName)
		}
	}
}

func getMapKeylist(mapList map[string]interface{}) *[]string {
	names := make([]string, 0)
	for key := range mapList {
		names = append(names, key)
	}
	return &names
}