package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewClusterCommand(t *testing.T) {
	cmd := NewClusterCommand()
	if cmd.Use != "cluster" {
		t.Errorf("Expected command use to be 'cluster', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected short description to be non-empty")
	}

	if len(cmd.Commands()) < 1 {
		t.Error("Expected cluster command to have subcommands")
	}

	createCommand := findSubcommand(cmd.Commands(), "create")
	if createCommand == nil {
		t.Error("Expected to find 'create' subcommand")
		return
	}

	nameFlag := createCommand.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("Expected create command to have 'name' flag")
	}

	resourceGroupFlag := createCommand.Flags().Lookup("resource-group")
	if resourceGroupFlag == nil {
		t.Error("Expected create command to have 'resource-group' flag")
	}
}

func TestCreateCommandFlagParsing(t *testing.T) {
	createCmd := newClusterCreateCommand()

	args := []string{
		"--name", "my-cluster",
		"--resource-group", "my-resource-group",
		"--location", "eastus2",
	}

	err := createCmd.ParseFlags(args)
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	name, _ := createCmd.Flags().GetString("name")
	if name != "my-cluster" {
		t.Errorf("Expected name flag to be 'my-cluster', got '%s'", name)
	}

	resourceGroup, _ := createCmd.Flags().GetString("resource-group")
	if resourceGroup != "my-resource-group" {
		t.Errorf("Expected resource-group flag to be 'my-resource-group', got '%s'", resourceGroup)
	}

	location, _ := createCmd.Flags().GetString("location")
	if location != "eastus2" {
		t.Errorf("Expected location flag to be 'eastus2', got '%s'", location)
	}

	nodeCount, _ := createCmd.Flags().GetInt("node-count")
	if nodeCount != 1 {
		t.Errorf("Expected node-count flag to be 1, got %d", nodeCount)
	}

	nodeVMSize, _ := createCmd.Flags().GetString("node-vm-size")
	if nodeVMSize != "Standard_DS2_v2" {
		t.Errorf("Expected node-vm-size flag to be 'Standard_DS2_v2', got '%s'", nodeVMSize)
	}
}

func TestCreateCommandCustomArgParsing(t *testing.T) {
	args := []string{
		"--name", "my-cluster",
		"--resource-group", "my-resource-group",
		"--location", "eastus2",
		"--node-count", "3",
		"--node-vm-size", "Standard_DS3_v2",
		"--skip-identity-creation",
		"--kubernetes-version", "1.23.5",
		"--load-balancer-sku", "standard",
		"--enable-managed-identity",
	}

	var name, resourceGroup, location, nodeVMSize string
	var skipIdentityCreation bool
	var nodeCount int
	customArgs := make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--name" || arg == "-n" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				name = args[i+1]
				i++
			}
		} else if arg == "--resource-group" || arg == "-g" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				resourceGroup = args[i+1]
				i++
			}
		} else if arg == "--location" || arg == "-l" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				location = args[i+1]
				i++
			}
		} else if arg == "--skip-identity-creation" {
			skipIdentityCreation = true
		} else if arg == "--node-count" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				count, err := strconv.Atoi(args[i+1])
				if err == nil {
					nodeCount = count
				}
				i++
			}
		} else if arg == "--node-vm-size" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				nodeVMSize = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--") {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				customArgs[arg] = args[i+1]
				i++
			} else {
				customArgs[arg] = ""
			}
		}
	}

	if name != "my-cluster" {
		t.Errorf("Expected name to be 'my-cluster', got '%s'", name)
	}

	if resourceGroup != "my-resource-group" {
		t.Errorf("Expected resourceGroup to be 'my-resource-group', got '%s'", resourceGroup)
	}

	if location != "eastus2" {
		t.Errorf("Expected location to be 'eastus2', got '%s'", location)
	}

	if nodeCount != 3 {
		t.Errorf("Expected nodeCount to be 3, got %d", nodeCount)
	}

	if nodeVMSize != "Standard_DS3_v2" {
		t.Errorf("Expected nodeVMSize to be 'Standard_DS3_v2', got '%s'", nodeVMSize)
	}

	if !skipIdentityCreation {
		t.Error("Expected skipIdentityCreation to be true, but it was false")
	}

	if val, exists := customArgs["--kubernetes-version"]; !exists || val != "1.23.5" {
		t.Errorf("Expected --kubernetes-version to be '1.23.5', got '%s'", val)
	}

	if val, exists := customArgs["--load-balancer-sku"]; !exists || val != "standard" {
		t.Errorf("Expected --load-balancer-sku to be 'standard', got '%s'", val)
	}

	if val, exists := customArgs["--enable-managed-identity"]; !exists || val != "" {
		t.Errorf("Expected --enable-managed-identity to be empty string, got '%s'", val)
	}
}

func findSubcommand(commands []*cobra.Command, name string) *cobra.Command {
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
