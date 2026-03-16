package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// Use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "627b10ce-37e2-41c0-bfe6-bd444b154bd9"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// Path to your Terraform code
		TerraformDir: "../",
		// Override default terraform variables
		Vars: map[string]interface{}{
			"label_prefix": "kass0112",
		},
	}

	// Clean up resources with `terraform destroy` at the end
	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`
	terraform.InitAndApply(t, terraformOptions)

	// Get Terraform outputs
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")

	// --- Test 1: Confirm VM exists ---
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// --- Test 2: Confirm NIC exists and is attached to VM ---
	assert.True(t, azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID))

	vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)

	// Dereference the pointer to get the slice
	nics := *vm.NetworkProfile.NetworkInterfaces
	if len(nics) == 0 {
		t.Fatal("VM has no network interfaces attached")
	}

	// Confirm one of the attached NICs matches the expected NIC
	foundNIC := false
	for _, nic := range nics {
		if *nic.ID != "" && (*nic.ID)[len(*nic.ID)-len(nicName):] == nicName {
			foundNIC = true
			break
		}
	}
	assert.True(t, foundNIC, "Expected NIC is not attached to the VM")

	// --- Test 3: Confirm VM is running correct Ubuntu version ---
	// Use ImageReference.Sku from VM storage profile
	expectedSKU := "22_04-lts-gen2"
	actualSKU := *vm.StorageProfile.ImageReference.Sku
	assert.Equal(t, expectedSKU, actualSKU, "VM is not running the expected Ubuntu version")
}
