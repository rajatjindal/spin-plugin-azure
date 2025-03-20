package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View and manage configuration settings for the Spin Azure CLI.`,
	}

	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigResetCommand())

	return cmd
}

func newConfigShowCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current configuration settings for the Spin Azure CLI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			switch outputFormat {
			case "json":
				jsonData, err := json.MarshalIndent(cfg, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal config to JSON: %w", err)
				}
				fmt.Println(string(jsonData))
			default:
				fmt.Println("Current Configuration:")
				fmt.Printf("  Subscription ID: %s\n", cfg.SubscriptionID)
				fmt.Printf("  Resource Group: %s\n", cfg.ResourceGroup)
				fmt.Printf("  Cluster Name: %s\n", cfg.ClusterName)
				fmt.Printf("  Location: %s\n", cfg.Location)
				fmt.Printf("  Identity Name: %s\n", cfg.IdentityName)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json)")

	return cmd
}

func newConfigResetCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  `Reset all configuration settings to their default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Print("Are you sure you want to reset all configuration? This will remove all saved settings. [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Reset cancelled.")
					return nil
				}
			}

			defaultConfig := &config.Config{}

			if err := config.SaveConfig(defaultConfig); err != nil {
				return fmt.Errorf("failed to reset config: %w", err)
			}

			fmt.Println("Configuration has been reset to defaults.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Force reset without confirmation")

	return cmd
}
