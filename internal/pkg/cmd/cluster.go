package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/aks"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage AKS clusters for Spin applications",
		Long:  `Create, use, and manage Azure Kubernetes Service (AKS) clusters for Spin applications.`,
	}

	cmd.AddCommand(newClusterCreateCommand())
	cmd.AddCommand(newClusterUseCommand())
	cmd.AddCommand(newClusterCheckIdentityCommand())
	cmd.AddCommand(newClusterInstallSpinOperatorCommand())

	return cmd
}

func newClusterCreateCommand() *cobra.Command {
	var name, resourceGroup, location, nodeVMSize string
	var nodeCount int
	var additionalArgs []string

	cmd := &cobra.Command{
		Use:                "create",
		Short:              "Create a new AKS cluster with workload identity enabled",
		Long:               `Create a new Azure Kubernetes Service (AKS) cluster with workload identity enabled and Spin Operator installed.`,
		DisableFlagParsing: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.DisableFlagParsing = true
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if resourceGroup == "" {
				return fmt.Errorf("--resource-group is required")
			}

			credential, err := config.GetAzureCredential()
			if err != nil {
				return fmt.Errorf("failed to get Azure credential: %w", err)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.SubscriptionID == "" {
				return fmt.Errorf("subscription ID not set, please set it using Azure CLI or environment variables")
			}

			aksService, err := aks.NewService(credential, cfg.SubscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create AKS service: %w", err)
			}

			if location == "" {
				location = "eastus"
			}

			if nodeCount <= 0 {
				nodeCount = 1
			}

			if nodeVMSize == "" {
				nodeVMSize = "Standard_DS2_v2"
			}

			for k, v := range customArgs {
				if v == "" {
					additionalArgs = append(additionalArgs, k)
				} else {
					additionalArgs = append(additionalArgs, k, v)
				}
			}

			fmt.Printf("Creating AKS cluster '%s' in resource group '%s' with %d nodes (VM size: %s)...\n",
				name, resourceGroup, nodeCount, nodeVMSize)
			if len(additionalArgs) > 0 {
				fmt.Println("Additional arguments passed to az aks create:", additionalArgs)
			}

			ctx := context.Background()
			if err := aksService.CreateCluster(ctx, resourceGroup, name, location, nodeCount, nodeVMSize, additionalArgs...); err != nil {
				return fmt.Errorf("failed to create AKS cluster: %w", err)
			}

			fmt.Printf("AKS cluster '%s' created successfully with workload identity enabled\n", name)

			fmt.Println("Installing Spin Operator (this may take a few minutes)...")
			if err := aksService.DeploySpinOperator(ctx); err != nil {
				return fmt.Errorf("failed to install Spin Operator: %w", err)
			}
			fmt.Println("Spin Operator installed successfully")

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the AKS cluster (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for the AKS cluster (required)")
	cmd.Flags().StringVar(&location, "location", "eastus", "Azure region for the AKS cluster")
	cmd.Flags().IntVar(&nodeCount, "node-count", 1, "Number of nodes in the AKS cluster")
	cmd.Flags().StringVar(&nodeVMSize, "node-vm-size", "Standard_DS2_v2", "VM size for the AKS cluster nodes")

	cmd.Long += `

  By default, no identity is created. Use 'spin-azure identity create' after creating the cluster.

  Any additional arguments provided will be passed directly to 'az aks create'.
  For example, you can specify '--kubernetes-version 1.23.5' to create a cluster with a specific Kubernetes version.
  
  See 'az aks create --help' for all available options.`

	return cmd
}

func newClusterUseCommand() *cobra.Command {
	var name, resourceGroup string
	var installSpinOperator bool

	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use an existing AKS cluster",
		Long:  `Configure the CLI to use an existing Azure Kubernetes Service (AKS) cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			credential, err := config.GetAzureCredential()
			if err != nil {
				return fmt.Errorf("failed to get Azure credential: %w", err)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.SubscriptionID == "" {
				return fmt.Errorf("subscription ID not set, please set it using Azure CLI or environment variables")
			}

			if resourceGroup == "" {
				if cfg.ResourceGroup == "" {
					return fmt.Errorf("resource group not set, please set it using --resource-group")
				}
				resourceGroup = cfg.ResourceGroup
			}

			aksService, err := aks.NewService(credential, cfg.SubscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create AKS service: %w", err)
			}

			fmt.Printf("Using existing AKS cluster '%s' in resource group '%s'...\n", name, resourceGroup)
			ctx := context.Background()
			if err := aksService.UseCluster(ctx, resourceGroup, name); err != nil {
				return fmt.Errorf("failed to use AKS cluster: %w", err)
			}

			fmt.Printf("Now using AKS cluster '%s'\n", name)

			if installSpinOperator {
				fmt.Println("Installing Spin Operator...")
				if err := aksService.DeploySpinOperator(ctx); err != nil {
					return fmt.Errorf("failed to install Spin Operator: %w", err)
				}
				fmt.Println("Spin Operator installed successfully")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the existing AKS cluster (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group of the existing AKS cluster")
	cmd.Flags().BoolVar(&installSpinOperator, "install-spin-operator", false, "Install Spin Operator on the cluster after selection")
	cmd.MarkFlagsRequiredTogether("name")
	return cmd
}

func newClusterCheckIdentityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-identity",
		Short: "Check if workload identity is enabled on the cluster",
		Long:  `Check if workload identity is enabled on the current AKS cluster, and enable it if not.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			credential, err := config.GetAzureCredential()
			if err != nil {
				return fmt.Errorf("failed to get Azure credential: %w", err)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.SubscriptionID == "" {
				return fmt.Errorf("subscription ID not set, please set it using Azure CLI or environment variables")
			}

			aksService, err := aks.NewService(credential, cfg.SubscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create AKS service: %w", err)
			}

			fmt.Println("Checking if workload identity is enabled on the cluster...")
			ctx := context.Background()
			enabled, err := aksService.CheckWorkloadIdentity(ctx)
			if err != nil {
				return fmt.Errorf("failed to check workload identity: %w", err)
			}

			if enabled {
				fmt.Println("Workload identity is already enabled on the cluster")
				return nil
			}

			fmt.Println("Workload identity is not enabled, enabling it now...")
			if err := aksService.EnableWorkloadIdentity(ctx); err != nil {
				return fmt.Errorf("failed to enable workload identity: %w", err)
			}

			fmt.Println("Workload identity has been enabled on the cluster")
			return nil
		},
	}

	return cmd
}

func newClusterInstallSpinOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install-spin-operator",
		Short: "Install Spin Operator on the current cluster",
		Long:  `Install Spin Operator and its dependencies on the current AKS cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			credential, err := config.GetAzureCredential()
			if err != nil {
				return fmt.Errorf("failed to get Azure credential: %w", err)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.SubscriptionID == "" {
				return fmt.Errorf("subscription ID not set, please set it using Azure CLI or environment variables")
			}

			aksService, err := aks.NewService(credential, cfg.SubscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create AKS service: %w", err)
			}

			fmt.Println("Installing Spin Operator on the current cluster...")
			ctx := context.Background()
			if err := aksService.DeploySpinOperator(ctx); err != nil {
				return fmt.Errorf("failed to install Spin Operator: %w", err)
			}

			fmt.Println("Spin Operator has been successfully installed on the cluster")
			return nil
		},
	}

	return cmd
}
