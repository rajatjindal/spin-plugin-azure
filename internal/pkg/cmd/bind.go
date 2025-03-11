package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/bind"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewBindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind Spin applications to Azure services",
		Long:  `Bind Spin applications to Azure services like CosmosDB.`,
	}

	cmd.AddCommand(newBindCosmosDBCommand())

	return cmd
}

func newBindCosmosDBCommand() *cobra.Command {
	var name, resourceGroup string

	cmd := &cobra.Command{
		Use:   "cosmosdb",
		Short: "Bind to an Azure CosmosDB instance",
		Long:  `Bind a Spin application to an Azure CosmosDB instance.`,
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

			cosmosDBService := bind.NewCosmosDBService(credential, cfg.SubscriptionID)

			fmt.Printf("Binding to CosmosDB '%s' in resource group '%s'...\n", name, resourceGroup)
			ctx := context.Background()
			if err := cosmosDBService.BindCosmosDB(ctx, name, resourceGroup); err != nil {
				return fmt.Errorf("failed to bind to CosmosDB: %w", err)
			}

			fmt.Printf("Successfully bound to CosmosDB '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the CosmosDB instance (required)")
	cmd.Flags().StringVar(&resourceGroup, "resource-group", "", "Resource group of the CosmosDB instance (required)")
	cmd.MarkFlagsRequiredTogether("name", "resource-group")

	return cmd
}
