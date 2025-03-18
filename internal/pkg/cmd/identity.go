package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/aks"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewIdentityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage Azure managed identities for Spin workloads",
		Long:  `Create and manage Azure managed identities and Kubernetes service accounts for Spin workloads.`,
	}

	cmd.AddCommand(newIdentityCreateCommand())

	return cmd
}

func newIdentityCreateCommand() *cobra.Command {
	var name, resourceGroup string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Azure managed identity with Kubernetes service account",
		Long:  `Create a new Azure managed identity and set up a Kubernetes service account with federated credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = "workload-identity"
			}

			// Load resource group from config if not provided
			if resourceGroup == "" {
				cfg, err := config.LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				if cfg.ResourceGroup == "" {
					return fmt.Errorf("--resource-group is required or use 'spin azure cluster use' to select a cluster first")
				}

				resourceGroup = cfg.ResourceGroup
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
				return fmt.Errorf("subscription ID not set, please set it using `spin azure login`")
			}

			aksService, err := aks.NewService(credential, cfg.SubscriptionID)
			if err != nil {
				return fmt.Errorf("failed to create AKS service: %w", err)
			}

			ctx := context.Background()

			fmt.Printf("Creating Azure managed identity '%s'...\n", name)
			if err := aksService.CreateIdentity(ctx, name, resourceGroup); err != nil {
				return fmt.Errorf("failed to create managed identity: %w", err)
			}

			fmt.Printf("Creating Kubernetes service account for identity '%s'...\n", name)
			if err := aksService.CreateServiceAccount(ctx, name); err != nil {
				return fmt.Errorf("failed to create service account: %w", err)
			}

			fmt.Printf("Identity and service account '%s' created successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "workload-identity", "Name of the identity to create")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group for the identity (defaults to the resource group of the current cluster)")
	cmd.MarkFlagRequired("name")
	return cmd
}
