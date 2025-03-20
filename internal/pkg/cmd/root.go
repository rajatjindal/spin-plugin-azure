package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Spin Azure CLI - Manage Spin apps on Azure Kubernetes Service (AKS)",
		Long: `A CLI tool for managing Spin apps on Azure Kubernetes Service (AKS).
This tool helps you create and manage AKS clusters with workload identities enabled,
and deploy Spin apps to them.`,
		Example: `  # Login to Azure
  spin azure login

  # Create a new AKS cluster
  spin azure cluster create --name my-cluster --resource-group my-rg --location eastus

  # Use an existing AKS cluster
  spin azure cluster use --name existing-cluster --resource-group existing-rg

  # Create a new identity
  spin azure identity create --name my-custom-identity

  # Create an identity without a Kubernetes cluster
  spin azure identity create --name my-identity --resource-group my-rg --skip-service-account

  # Use an existing identity and create a service account
  spin azure identity use --name my-custom-identity --create-service-account

  # Assign Azure CosmosDB role to an identity
  spin azure assign-role cosmosdb --name my-cosmos --resource-group my-rg

  # Deploy a Spin application
  spin azure deploy --from path/to/spinapp.yaml
 
  # Output the config
  spin azure config show

  # Reset the config
  spin azure config reset -y`,
	}

	cmd.AddCommand(NewLoginCommand())
	cmd.AddCommand(NewLogoutCommand())
	cmd.AddCommand(NewClusterCommand())
	cmd.AddCommand(NewIdentityCommand())
	cmd.AddCommand(NewAssignRoleCommand())
	cmd.AddCommand(NewDeployCommand())
	cmd.AddCommand(NewConfigCommand())

	return cmd
}
