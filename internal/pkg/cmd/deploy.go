package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/deploy"
)

// NewDeployCommand creates a new deploy command
func NewDeployCommand() *cobra.Command {
	var from, identity string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Spin applications to AKS",
		Long:  `Deploy Spin applications to Azure Kubernetes Service (AKS) with workload identity.`,
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

			if from == "" {
				return fmt.Errorf("--from flag is required, please specify the path to the SpinApp YAML file")
			}

			if identity == "" {
				identity = "workload-identity"
				fmt.Println("No identity specified, using default 'workload-identity'")
			}

			deployService := deploy.NewService(credential, cfg.SubscriptionID)

			fmt.Printf("Deploying Spin application from '%s' using identity '%s'...\n", from, identity)
			ctx := context.Background()
			if err := deployService.Deploy(ctx, from, identity); err != nil {
				return fmt.Errorf("failed to deploy Spin application: %w", err)
			}

			fmt.Printf("Successfully deployed Spin application from '%s'\n", from)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Path to the SpinApp YAML file (required)")
	cmd.Flags().StringVar(&identity, "identity", "workload-identity", "Name of the workload identity to use (default: workload-identity)")
	cmd.MarkFlagRequired("from")

	return cmd
}
