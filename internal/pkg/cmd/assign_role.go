package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/bind"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewAssignRoleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-role",
		Short: "Assign Azure RBAC roles to managed identities",
		Long:  `Assign Azure RBAC roles to managed identities for accessing Azure services like CosmosDB.`,
	}

	cmd.AddCommand(newBindCosmosDBCommand())

	return cmd
}

func newBindCosmosDBCommand() *cobra.Command {
	var name, resourceGroup string

	cmd := &cobra.Command{
		Use:   "cosmosdb",
		Short: "Assign Azure roles for CosmosDB access",
		Long:  `Assign the necessary Azure RBAC roles to a managed identity for accessing an Azure CosmosDB instance.`,
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

			cosmosDBService := bind.NewCosmosDBService(credential, cfg.SubscriptionID)

			fmt.Printf("Assigning CosmosDB Data Contributor role to identity '%s' for CosmosDB account '%s' in resource group '%s'...\n", cfg.WorkloadIdentity, name, resourceGroup)
			ctx := context.Background()
			if err := cosmosDBService.BindCosmosDB(ctx, name, resourceGroup); err != nil {
				return fmt.Errorf("failed to assign role to CosmosDB: %w", err)
			}

			fmt.Printf("Successfully assigned roles to CosmosDB '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the CosmosDB account (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group of the CosmosDB account")
	cmd.MarkFlagsRequiredTogether("name")

	return cmd
}
