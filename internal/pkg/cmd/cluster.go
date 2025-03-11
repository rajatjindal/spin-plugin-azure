package cmd

import (
	"context"
	"fmt"

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
	var name, resourceGroup, location, createIdentity string
	var skipIdentityCreation bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new AKS cluster with workload identity enabled",
		Long:  `Create a new Azure Kubernetes Service (AKS) cluster with workload identity enabled, Spin Operator installed, and a managed identity with service account configured.`,
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

			if location == "" {
				location = "eastus"
			}

			fmt.Printf("Creating AKS cluster '%s' in resource group '%s'...\n", name, resourceGroup)
			ctx := context.Background()
			if err := aksService.CreateCluster(ctx, resourceGroup, name, location); err != nil {
				return fmt.Errorf("failed to create AKS cluster: %w", err)
			}

			fmt.Printf("AKS cluster '%s' created successfully with workload identity enabled\n", name)

			fmt.Println("Installing Spin Operator (this may take a few minutes)...")
			if err := aksService.DeploySpinOperator(ctx); err != nil {
				return fmt.Errorf("failed to install Spin Operator: %w", err)
			}
			fmt.Println("Spin Operator installed successfully")

			if !skipIdentityCreation {
				identityName := createIdentity
				if identityName == "" {
					identityName = "workload-identity"
				}

				fmt.Printf("Creating Azure managed identity '%s'...\n", identityName)
				if err := aksService.CreateIdentity(ctx, identityName, resourceGroup); err != nil {
					return fmt.Errorf("failed to create managed identity: %w", err)
				}

				fmt.Printf("Creating Kubernetes service account for identity '%s'...\n", identityName)
				if err := aksService.CreateServiceAccount(ctx, identityName); err != nil {
					return fmt.Errorf("failed to create service account: %w", err)
				}

				fmt.Printf("Identity and service account '%s' created successfully\n", identityName)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the AKS cluster (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for the AKS cluster (required)")
	cmd.Flags().StringVar(&location, "location", "", "Azure region for the AKS cluster (default: eastus)")
	cmd.Flags().StringVar(&createIdentity, "create-identity", "workload-identity", "Name of the identity to create (default: workload-identity)")
	cmd.Flags().BoolVar(&skipIdentityCreation, "skip-identity-creation", false, "Skip creation of managed identity and service account")
	cmd.MarkFlagsRequiredTogether("name", "resource-group")

	return cmd
}

func newClusterUseCommand() *cobra.Command {
	var name, resourceGroup, createIdentity string
	var installSpinOperator, skipIdentityCreation bool

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

			if !skipIdentityCreation && createIdentity != "" {
				fmt.Printf("Creating Azure managed identity '%s'...\n", createIdentity)
				if err := aksService.CreateIdentity(ctx, createIdentity, resourceGroup); err != nil {
					return fmt.Errorf("failed to create managed identity: %w", err)
				}

				fmt.Printf("Creating Kubernetes service account for identity '%s'...\n", createIdentity)
				if err := aksService.CreateServiceAccount(ctx, createIdentity); err != nil {
					return fmt.Errorf("failed to create service account: %w", err)
				}

				fmt.Printf("Identity and service account '%s' created successfully\n", createIdentity)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the existing AKS cluster (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group of the existing AKS cluster (required)")
	cmd.Flags().BoolVar(&installSpinOperator, "install-spin-operator", false, "Install Spin Operator on the cluster after selection")
	cmd.Flags().StringVar(&createIdentity, "create-identity", "", "Name of the identity to create")
	cmd.Flags().BoolVar(&skipIdentityCreation, "skip-identity-creation", false, "Skip creation of managed identity and service account")
	cmd.MarkFlagsRequiredTogether("name", "resource-group")

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
