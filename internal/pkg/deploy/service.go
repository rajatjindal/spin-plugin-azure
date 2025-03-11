package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

type Service struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func NewService(credential azcore.TokenCredential, subscriptionID string) *Service {
	return &Service{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

func (s *Service) Deploy(ctx context.Context, spinAppYAMLPath, identityName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ClusterName == "" || cfg.ResourceGroup == "" {
		return fmt.Errorf("no cluster is currently selected, use 'spin-azure cluster use' or 'spin-azure cluster create' first")
	}

	if _, err := os.Stat(spinAppYAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("SpinApp YAML file not found at %s", spinAppYAMLPath)
	}

	// Get Kubernetes credentials for the current cluster
	if err := s.getKubernetesCredentials(cfg.ClusterName, cfg.ResourceGroup); err != nil {
		return err
	}

	// Verify service account exists
	namespace := "default"
	checkCmd := exec.Command("kubectl", "get", "serviceaccount", identityName, "-n", namespace, "--ignore-not-found")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check if service account exists: %w\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), identityName) {
		return fmt.Errorf("service account '%s' not found in namespace '%s', please create it using 'spin-azure cluster use --service-account=%s' or 'spin-azure cluster create --service-account=%s'", identityName, namespace, identityName, identityName)
	}

	if err := s.deploySpinAppYAML(spinAppYAMLPath); err != nil {
		return err
	}

	fmt.Printf("Successfully deployed Spin application from '%s' with identity '%s'\n", spinAppYAMLPath, identityName)
	return nil
}

func (s *Service) getIdentityClientID(name, resourceGroup string) (string, error) {
	cmd := exec.Command(
		"az", "identity", "show",
		"--name", name,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--query", "clientId",
		"--output", "tsv",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get identity client ID: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func (s *Service) getKubernetesCredentials(clusterName, resourceGroup string) error {
	cmd := exec.Command(
		"az", "aks", "get-credentials",
		"--name", clusterName,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--overwrite-existing",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes credentials: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *Service) deploySpinAppYAML(spinAppYAMLPath string) error {
	checkCmd := exec.Command("kubectl", "apply", "--dry-run=client", "-f", spinAppYAMLPath, "-o", "name")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to parse YAML file: %w\nOutput: %s", err, string(output))
	}

	resourceNames := strings.Split(string(output), "\n")
	var spinAppName string
	for _, name := range resourceNames {
		if strings.Contains(name, "spinapp") {
			parts := strings.Split(name, "/")
			if len(parts) > 1 {
				spinAppName = parts[1]
			}
			break
		}
	}

	if spinAppName != "" {
		fmt.Printf("Deploying SpinApp '%s'\n", spinAppName)
	} else {
		fmt.Println("Deploying SpinApp resources")
	}

	cmd := exec.Command("kubectl", "apply", "-f", spinAppYAMLPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply SpinApp: %w\nOutput: %s", err, string(output))
	}

	if spinAppName != "" {
		fmt.Printf("SpinApp '%s' deployed successfully\n", spinAppName)
	} else {
		fmt.Println("SpinApp resources deployed successfully")
	}

	return nil
}
