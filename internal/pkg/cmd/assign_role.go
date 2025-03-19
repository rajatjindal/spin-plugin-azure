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
	var name, resourceGroup, identityName, identityResourceGroup string

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
				return fmt.Errorf("subscription ID not set, please set it using `spin azure login`")
			}

			if resourceGroup == "" {

				resourceGroup = cfg.ResourceGroup
			}

			if resourceGroup == "" {
				return fmt.Errorf("resource group for CosmosDB not set, please set it using --resource-group")
			}

			if identityResourceGroup == "" {
				identityResourceGroup = resourceGroup
			}

			if identityName == "" {
				identityName = cfg.IdentityName
			}

			if identityName == "" {
				return fmt.Errorf("identity name not set, please set it using --identity")
			}

			cosmosDBService := bind.NewCosmosDBService(credential, cfg.SubscriptionID)

			fmt.Printf("Assigning CosmosDB Data Contributor role to identity '%s' (in resource group '%s') for CosmosDB account '%s' (in resource group '%s')...\n",
				identityName, identityResourceGroup, name, resourceGroup)

			ctx := context.Background()
			if err := cosmosDBService.BindCosmosDB(ctx, name, resourceGroup, identityName, identityResourceGroup); err != nil {
				return fmt.Errorf("failed to assign role to CosmosDB: %w", err)
			}

			fmt.Printf("Successfully assigned roles to CosmosDB '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the CosmosDB account (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group of the CosmosDB account")
	cmd.Flags().StringVar(&identityName, "identity", "", "Name of the identity to assign roles to")
	cmd.Flags().StringVar(&identityResourceGroup, "identity-resource-group", "", "Resource group of the managed identity (defaults to the CosmosDB resource group if not specified)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Sprintf("failed to mark flag 'from' as required: %v", err))
	}

	return cmd
}
