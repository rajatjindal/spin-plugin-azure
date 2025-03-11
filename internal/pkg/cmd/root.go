package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spin-azure",
		Short: "Spin Azure CLI - Manage Spin apps on Azure Kubernetes Service (AKS)",
		Long: `A CLI tool for deploying and managing Spin applications on Azure Kubernetes Service (AKS).
This tool helps you create and manage AKS clusters, set up workload identities,
bind to Azure services like CosmosDB, and deploy Spin applications to AKS.`,
	}

	cmd.AddCommand(NewLoginCommand())
	cmd.AddCommand(NewLogoutCommand())
	cmd.AddCommand(NewClusterCommand())
	cmd.AddCommand(NewBindCommand())
	cmd.AddCommand(NewDeployCommand())

	return cmd
}
