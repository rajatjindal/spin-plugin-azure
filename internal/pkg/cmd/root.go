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
	}

	cmd.AddCommand(NewLoginCommand())
	cmd.AddCommand(NewLogoutCommand())
	cmd.AddCommand(NewClusterCommand())
	cmd.AddCommand(NewBindCommand())
	cmd.AddCommand(NewDeployCommand())

	return cmd
}
