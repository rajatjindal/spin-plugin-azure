package bind

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/spinframework/spin-plugin-azure/internal/pkg/config"
)

type CosmosDBService struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func NewCosmosDBService(credential azcore.TokenCredential, subscriptionID string) *CosmosDBService {
	return &CosmosDBService{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

func (s *CosmosDBService) BindCosmosDB(ctx context.Context, name, resourceGroup string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.WorkloadIdentity == "" {
		return fmt.Errorf("no workload identity configured, use 'spin-azure cluster' first")
	}

	if err := s.validateCosmosDBAccount(name, resourceGroup); err != nil {
		return err
	}

	identityPrincipalID, err := s.getIdentityPrincipalID(cfg.WorkloadIdentity, cfg.ResourceGroup)
	if err != nil {
		return err
	}

	if err := s.assignRoleToCosmosDB(identityPrincipalID, name, resourceGroup); err != nil {
		return err
	}

	fmt.Printf("Successfully bound CosmosDB '%s' to identity '%s'\n", name, cfg.WorkloadIdentity)

	dbName, containerName, err := s.getDBAndContainerInfo(name, resourceGroup)
	if err == nil && dbName != "" && containerName != "" {
		fmt.Printf("CosmosDB Database: %s, Container: %s\n", dbName, containerName)
	}

	return nil
}

func (s *CosmosDBService) validateCosmosDBAccount(name, resourceGroup string) error {
	cmd := exec.Command(
		"az", "cosmosdb", "check-name-exists",
		"--name", name,
		"--subscription", s.subscriptionID,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check if CosmosDB exists: %w\nOutput: %s", err, string(output))
	}

	cmd = exec.Command(
		"az", "cosmosdb", "show",
		"--name", name,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CosmosDB '%s' not found in resource group '%s': %w\nOutput: %s",
			name, resourceGroup, err, string(output))
	}

	return nil
}

func (s *CosmosDBService) getIdentityPrincipalID(name, resourceGroup string) (string, error) {
	cmd := exec.Command(
		"az", "identity", "show",
		"--name", name,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--query", "principalId",
		"--output", "tsv",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get identity principal ID: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func (s *CosmosDBService) assignRoleToCosmosDB(identityPrincipalID, cosmosDBName, resourceGroup string) error {
	cosmosDBResourceID, err := s.getCosmosDBResourceID(cosmosDBName, resourceGroup)
	if err != nil {
		return err
	}

	roleDefinitionID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DocumentDB/databaseAccounts/%s/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002", s.subscriptionID, resourceGroup, cosmosDBName)

	cmd := exec.Command(
		"az", "cosmosdb", "sql", "role", "assignment", "create",
		"--account-name", cosmosDBName,
		"--resource-group", resourceGroup,
		"--role-definition-id", roleDefinitionID,
		"--principal-id", identityPrincipalID,
		"--scope", cosmosDBResourceID,
		"--subscription", s.subscriptionID,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to assign role to CosmosDB: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (s *CosmosDBService) getCosmosDBResourceID(name, resourceGroup string) (string, error) {
	cmd := exec.Command(
		"az", "cosmosdb", "show",
		"--name", name,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--query", "id",
		"--output", "tsv",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get CosmosDB resource ID: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func (s *CosmosDBService) getDBAndContainerInfo(name, resourceGroup string) (string, string, error) {
	dbCmd := exec.Command(
		"az", "cosmosdb", "sql", "database", "list",
		"--account-name", name,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--query", "[0].name",
		"--output", "tsv",
	)

	dbOutput, err := dbCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to get database info: %w", err)
	}

	dbName := strings.TrimSpace(string(dbOutput))
	if dbName == "" {
		return "", "", fmt.Errorf("no SQL database found in account %s", name)
	}

	containerCmd := exec.Command(
		"az", "cosmosdb", "sql", "container", "list",
		"--account-name", name,
		"--database-name", dbName,
		"--resource-group", resourceGroup,
		"--subscription", s.subscriptionID,
		"--query", "[0].name",
		"--output", "tsv",
	)

	containerOutput, err := containerCmd.CombinedOutput()
	if err != nil {
		return dbName, "", fmt.Errorf("failed to get container info: %w", err)
	}

	containerName := strings.TrimSpace(string(containerOutput))
	return dbName, containerName, nil
}
