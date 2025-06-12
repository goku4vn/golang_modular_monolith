package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// VaultConfig holds Vault-specific configuration
type VaultConfig struct {
	Address    string `mapstructure:"address"`
	Token      string `mapstructure:"token"`
	RoleID     string `mapstructure:"role_id"`
	SecretID   string `mapstructure:"secret_id"`
	MountPath  string `mapstructure:"mount_path"`
	SecretPath string `mapstructure:"secret_path"`
	Enabled    bool   `mapstructure:"enabled"`
}

// VaultClient wraps the Vault API client
type VaultClient struct {
	client *api.Client
	config VaultConfig
}

// NewVaultClient creates a new Vault client
func NewVaultClient() (*VaultClient, error) {
	config := VaultConfig{
		Address:    getEnvOrDefault("VAULT_ADDR", "http://localhost:8200"),
		Token:      os.Getenv("VAULT_TOKEN"),
		RoleID:     os.Getenv("VAULT_ROLE_ID"),
		SecretID:   os.Getenv("VAULT_SECRET_ID"),
		MountPath:  getEnvOrDefault("VAULT_MOUNT_PATH", "secret"),
		SecretPath: getEnvOrDefault("VAULT_SECRET_PATH", "modular-monolith"),
		Enabled:    getEnvOrDefault("VAULT_ENABLED", "false") == "true",
	}

	if !config.Enabled {
		log.Println("🔒 Vault is disabled, skipping Vault client initialization")
		return &VaultClient{config: config}, nil
	}

	// Create Vault client
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = config.Address

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	vaultClient := &VaultClient{
		client: client,
		config: config,
	}

	// Authenticate with Vault
	if err := vaultClient.authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate with Vault: %w", err)
	}

	log.Println("🔒 Vault client initialized successfully")
	return vaultClient, nil
}

// authenticate handles Vault authentication
func (vc *VaultClient) authenticate() error {
	if vc.config.Token != "" {
		// Use token authentication
		vc.client.SetToken(vc.config.Token)
		log.Println("🔑 Using Vault token authentication")
		return nil
	}

	if vc.config.RoleID != "" && vc.config.SecretID != "" {
		// Use AppRole authentication
		return vc.authenticateWithAppRole()
	}

	return fmt.Errorf("no valid authentication method found (token or AppRole)")
}

// authenticateWithAppRole authenticates using AppRole method
func (vc *VaultClient) authenticateWithAppRole() error {
	data := map[string]interface{}{
		"role_id":   vc.config.RoleID,
		"secret_id": vc.config.SecretID,
	}

	resp, err := vc.client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return fmt.Errorf("AppRole authentication failed: %w", err)
	}

	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned from AppRole login")
	}

	vc.client.SetToken(resp.Auth.ClientToken)
	log.Println("🔑 AppRole authentication successful")

	// Set up token renewal
	go vc.renewToken(resp.Auth.ClientToken, time.Duration(resp.Auth.LeaseDuration)*time.Second)

	return nil
}

// renewToken handles automatic token renewal
func (vc *VaultClient) renewToken(token string, leaseDuration time.Duration) {
	ticker := time.NewTicker(leaseDuration / 2) // Renew at half the lease duration
	defer ticker.Stop()

	for range ticker.C {
		resp, err := vc.client.Auth().Token().RenewSelf(0)
		if err != nil {
			log.Printf("❌ Failed to renew Vault token: %v", err)
			continue
		}

		if resp.Auth != nil {
			leaseDuration = time.Duration(resp.Auth.LeaseDuration) * time.Second
			ticker.Reset(leaseDuration / 2)
			log.Println("🔄 Vault token renewed successfully")
		}
	}
}

// LoadSecrets loads secrets from Vault and sets them in Viper
func (vc *VaultClient) LoadSecrets(modulesConfig *ModulesConfig) error {
	if !vc.config.Enabled || vc.client == nil {
		log.Println("🔒 Vault is disabled, skipping secret loading")
		return nil
	}

	totalSecrets := 0

	// Load app-level secrets
	if err := vc.loadSecretsFromPath("app", "app"); err != nil {
		log.Printf("⚠️ Failed to load app secrets: %v", err)
	} else {
		count, _ := vc.getSecretCount("app")
		totalSecrets += count
		log.Printf("📱 Loaded %d app secrets", count)
	}

	// Load module secrets dynamically from configuration
	if modulesConfig != nil {
		for moduleName, moduleConfig := range modulesConfig.Modules {
			if moduleConfig.Vault.Enabled {
				modulePath := moduleConfig.Vault.Path
				if err := vc.loadSecretsFromPath(modulePath, moduleName); err != nil {
					log.Printf("⚠️ Failed to load %s module secrets: %v", moduleName, err)
				} else {
					count, _ := vc.getSecretCount(modulePath)
					totalSecrets += count
					log.Printf("🔧 Loaded %d secrets for %s module", count, moduleName)
				}
			} else {
				log.Printf("🔒 Vault disabled for module: %s", moduleName)
			}
		}
	} else {
		// No modules config available, skip module secrets loading
		log.Println("⚠️ No modules config available, skipping module secrets loading")
	}

	log.Printf("🔒 Total loaded %d secrets from Vault", totalSecrets)
	return nil
}

// loadSecretsFromPath loads secrets from a specific Vault path
func (vc *VaultClient) loadSecretsFromPath(vaultPath, module string) error {
	secretPath := fmt.Sprintf("%s/data/%s", vc.config.MountPath, vaultPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secret, err := vc.client.Logical().ReadWithContext(ctx, secretPath)
	if err != nil {
		return fmt.Errorf("failed to read secret from path %s: %w", secretPath, err)
	}

	if secret == nil {
		return fmt.Errorf("no secret found at path: %s", secretPath)
	}

	// Extract data from KV v2 format
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid secret format at path %s", secretPath)
	}

	// Set secrets in Viper with high priority
	for key, value := range data {
		if strValue, ok := value.(string); ok {
			// Convert Vault key format to Viper format based on module
			viperKey := vc.convertModuleKeyToViperKey(key, module)
			viper.Set(viperKey, strValue)
		}
	}

	return nil
}

// getSecretCount returns the number of secrets at a path
func (vc *VaultClient) getSecretCount(vaultPath string) (int, error) {
	secretPath := fmt.Sprintf("%s/data/%s", vc.config.MountPath, vaultPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	secret, err := vc.client.Logical().ReadWithContext(ctx, secretPath)
	if err != nil || secret == nil {
		return 0, err
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return 0, nil
	}

	return len(data), nil
}

// convertModuleKeyToViperKey converts Vault key format to Viper nested key format based on module
func (vc *VaultClient) convertModuleKeyToViperKey(vaultKey, module string) string {
	key := strings.ToLower(vaultKey)

	// Handle app module (app-level configs)
	if module == "app" {
		switch key {
		case "app_version":
			return "app.version"
		case "app_name":
			return "app.name"
		case "gin_mode":
			return "app.gin_mode"
		case "port":
			return "app.port"
		default:
			return fmt.Sprintf("app.%s", key)
		}
	}

	// Handle database keys for modules
	if strings.HasPrefix(key, "database_") {
		field := strings.TrimPrefix(key, "database_")
		return fmt.Sprintf("databases.%s.%s", module, field)
	}

	// Handle module-specific keys (store in module namespace)
	return fmt.Sprintf("modules.%s.%s", module, key)
}

// convertVaultKeyToViperKey converts Vault key format to Viper nested key format (legacy method)
func (vc *VaultClient) convertVaultKeyToViperKey(vaultKey string) string {
	// Convert CUSTOMER_DATABASE_HOST to databases.customer.host
	// Convert APP_VERSION to app.version

	key := strings.ToLower(vaultKey)

	// Handle database keys
	if strings.Contains(key, "_database_") {
		parts := strings.Split(key, "_database_")
		if len(parts) == 2 {
			module := parts[0]
			field := parts[1]
			return fmt.Sprintf("databases.%s.%s", module, field)
		}
	}

	// Handle app keys
	if strings.HasPrefix(key, "app_") {
		field := strings.TrimPrefix(key, "app_")
		return fmt.Sprintf("app.%s", field)
	}

	// Handle special cases
	switch key {
	case "gin_mode":
		return "app.gin_mode"
	case "port":
		return "app.port"
	case "app_version":
		return "app.version"
	}

	// Default: convert underscores to dots
	return strings.ReplaceAll(key, "_", ".")
}

// IsEnabled returns true if Vault is enabled
func (vc *VaultClient) IsEnabled() bool {
	return vc.config.Enabled
}

// GetConfig returns the Vault configuration
func (vc *VaultClient) GetConfig() VaultConfig {
	return vc.config
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
