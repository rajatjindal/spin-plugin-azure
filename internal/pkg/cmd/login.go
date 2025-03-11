package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewLoginCommand() *cobra.Command {
	var subscriptionID string
	var tenantID string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Azure",
		Long:  `Log in to Azure and configure the CLI to use your Azure account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Logging in to Azure...")
			loginCmd := exec.Command("az", "login")
			loginCmd.Stdin = cmd.InOrStdin()
			loginCmd.Stdout = cmd.OutOrStdout()
			loginCmd.Stderr = cmd.ErrOrStderr()

			if err := loginCmd.Run(); err != nil {
				return fmt.Errorf("failed to log in to Azure: %w", err)
			}

			if subscriptionID != "" {
				fmt.Printf("Setting subscription to '%s'...\n", subscriptionID)
				subCmd := exec.Command("az", "account", "set", "--subscription", subscriptionID)

				output, err := subCmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("failed to set subscription: %w\nOutput: %s", err, string(output))
				}
			} else {
				subCmd := exec.Command("az", "account", "show", "--query", "id", "--output", "tsv")
				output, err := subCmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("failed to get current subscription: %w\nOutput: %s", err, string(output))
				}
				subscriptionID = strings.TrimSpace(string(output))
				fmt.Printf("Using subscription: %s\n", subscriptionID)
			}

			if tenantID != "" {
				fmt.Printf("Using tenant ID: %s\n", tenantID)
			} else {
				tenantCmd := exec.Command("az", "account", "show", "--query", "tenantId", "--output", "tsv")
				output, err := tenantCmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("failed to get current tenant: %w\nOutput: %s", err, string(output))
				}
				tenantID = strings.TrimSpace(string(output))
				fmt.Printf("Using tenant ID: %s\n", tenantID)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cfg.SubscriptionID = subscriptionID
			cfg.TenantID = tenantID

			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			verifyCmd := exec.Command("az", "account", "list", "--output", "none")
			if err := verifyCmd.Run(); err != nil {
				return fmt.Errorf("login verification failed: %w", err)
			}

			fmt.Println("Successfully logged in to Azure!")
			return nil
		},
	}

	cmd.Flags().StringVar(&subscriptionID, "subscription", "", "Azure subscription ID to use")
	cmd.Flags().StringVar(&tenantID, "tenant", "", "Azure tenant ID to use")

	return cmd
}

func NewLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from Azure",
		Long:  `Log out from Azure and clear saved credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Logging out from Azure...")
			logoutCmd := exec.Command("az", "logout")

			output, err := logoutCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to log out from Azure: %w\nOutput: %s", err, string(output))
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cfg.SubscriptionID = ""
			cfg.TenantID = ""

			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("Successfully logged out from Azure!")
			return nil
		},
	}

	return cmd
}

func azureTokenScope() string {
	return "https://management.azure.com/.default"
}
