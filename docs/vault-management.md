# Vault Management

Hướng dẫn quản lý HashiCorp Vault trong Modular Monolith để secure secrets và configuration.

## Overview

Modular Monolith sử dụng **HashiCorp Vault** để:
- **Secure Secret Storage**: Database passwords, API keys, certificates
- **Dynamic Configuration**: Runtime configuration management
- **Environment Isolation**: Different secrets per environment
- **Audit Logging**: Track secret access và modifications

## Vault Architecture

### Development Mode
```
┌─────────────────────────────────────┐
│            Vault Container          │
│         (Development Mode)          │
├─────────────────────────────────────┤
│  • In-memory storage               │
│  • Auto-unsealed                   │
│  • Root token: dev-root-token      │
│  • HTTP API: localhost:8200        │
└─────────────────────────────────────┘
```

### Production Mode
```
┌─────────────────────────────────────┐
│            Vault Cluster            │
│         (Production Mode)           │
├─────────────────────────────────────┤
│  • Persistent storage              │
│  • Manual unsealing required       │
│  • TLS encryption                  │
│  • High availability               │
└─────────────────────────────────────┘
```

## Vault Setup

### Development Environment

#### 1. Start Vault Container
```bash
# Vault starts automatically with docker-dev
make docker-dev

# Or start Vault separately
docker-compose -f docker/docker-compose.dev.yml up vault
```

#### 2. Verify Vault Status
```bash
# Check container status
docker ps | grep vault

# Check Vault health
curl http://localhost:8200/v1/sys/health
```

#### 3. Access Vault UI
```bash
# Open Vault UI in browser
open http://localhost:8200

# Login with development token
Token: dev-root-token
```

### Production Environment

#### 1. Initialize Vault
```bash
# Initialize Vault cluster
vault operator init

# Example output:
# Unseal Key 1: ABC123...
# Unseal Key 2: DEF456...
# Unseal Key 3: GHI789...
# Initial Root Token: s.XYZ789...
```

#### 2. Unseal Vault
```bash
# Unseal with 3 keys (threshold)
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>
```

#### 3. Configure Authentication
```bash
# Enable AppRole authentication
vault auth enable approle

# Create policy for application
vault policy write tmm-policy - <<EOF
path "secret/data/tmm/*" {
  capabilities = ["read"]
}
EOF

# Create AppRole
vault write auth/approle/role/tmm \
    token_policies="tmm-policy" \
    token_ttl=1h \
    token_max_ttl=4h
```

## Secret Management

### Secret Structure
```
secret/
├── tmm/                    # Application namespace
│   ├── development/        # Development environment
│   │   ├── database/      # Database credentials
│   │   ├── redis/         # Redis credentials
│   │   └── external/      # External API keys
│   ├── staging/           # Staging environment
│   └── production/        # Production environment
```

### Storing Secrets

#### Database Credentials
```bash
# Store database secrets
vault kv put secret/tmm/development/database \
    host="localhost" \
    port="5432" \
    username="postgres" \
    password="secure-password" \
    database="tmm_customer"

# Store for different modules
vault kv put secret/tmm/development/database/customer \
    host="customer-db.example.com" \
    username="customer_user" \
    password="customer-password"

vault kv put secret/tmm/development/database/order \
    host="order-db.example.com" \
    username="order_user" \
    password="order-password"
```

#### External API Keys
```bash
# Store external service credentials
vault kv put secret/tmm/development/external \
    stripe_api_key="sk_test_..." \
    sendgrid_api_key="SG...." \
    aws_access_key="AKIA..." \
    aws_secret_key="..."

# Store OAuth credentials
vault kv put secret/tmm/development/oauth \
    google_client_id="..." \
    google_client_secret="..." \
    github_client_id="..." \
    github_client_secret="..."
```

#### Application Configuration
```bash
# Store application-specific config
vault kv put secret/tmm/development/app \
    jwt_secret="super-secret-key" \
    encryption_key="32-byte-encryption-key" \
    session_secret="session-secret-key"
```

### Reading Secrets

#### Using Vault CLI
```bash
# Read database secrets
vault kv get secret/tmm/development/database

# Read specific field
vault kv get -field=password secret/tmm/development/database

# Read in JSON format
vault kv get -format=json secret/tmm/development/database
```

#### Using HTTP API
```bash
# Get secrets via API
curl -H "X-Vault-Token: $VAULT_TOKEN" \
     http://localhost:8200/v1/secret/data/tmm/development/database

# Response format:
{
  "data": {
    "data": {
      "host": "localhost",
      "password": "secure-password",
      "port": "5432",
      "username": "postgres"
    }
  }
}
```

## Application Integration

### Vault Configuration

#### Environment Variables
```bash
# Vault connection settings
export VAULT_ENABLED=true
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token
export VAULT_SECRET_PATH=tmm
export VAULT_ENVIRONMENT=development
```

#### Module Configuration
```yaml
# internal/modules/customer/module.yaml
enabled: true
vault:
  enabled: true
  secret_path: "secret/tmm/development/database/customer"
database:
  host: "${VAULT:host}"
  port: "${VAULT:port}"
  username: "${VAULT:username}"
  password: "${VAULT:password}"
  name: "tmm_customer"
```

### Go Integration

#### Vault Client Setup
```go
// internal/shared/infrastructure/vault/client.go
package vault

import (
    "github.com/hashicorp/vault/api"
)

type Client struct {
    client *api.Client
    token  string
}

func NewClient(addr, token string) (*Client, error) {
    config := api.DefaultConfig()
    config.Address = addr
    
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    client.SetToken(token)
    
    return &Client{
        client: client,
        token:  token,
    }, nil
}

func (c *Client) GetSecret(path string) (map[string]interface{}, error) {
    secret, err := c.client.Logical().Read(path)
    if err != nil {
        return nil, err
    }
    
    if secret == nil || secret.Data == nil {
        return nil, fmt.Errorf("secret not found at path: %s", path)
    }
    
    // Handle KV v2 format
    if data, ok := secret.Data["data"].(map[string]interface{}); ok {
        return data, nil
    }
    
    return secret.Data, nil
}
```

#### Configuration Loading
```go
// internal/shared/infrastructure/config/vault.go
func LoadVaultConfig(vaultClient *vault.Client, secretPath string) (map[string]string, error) {
    secrets, err := vaultClient.GetSecret(secretPath)
    if err != nil {
        return nil, err
    }
    
    config := make(map[string]string)
    for key, value := range secrets {
        if str, ok := value.(string); ok {
            config[key] = str
        }
    }
    
    return config, nil
}

// Replace ${VAULT:key} placeholders
func ReplaceVaultPlaceholders(config string, vaultSecrets map[string]string) string {
    for key, value := range vaultSecrets {
        placeholder := fmt.Sprintf("${VAULT:%s}", key)
        config = strings.ReplaceAll(config, placeholder, value)
    }
    return config
}
```

## Environment-Specific Configuration

### Development Environment
```bash
# Development secrets (less secure, easier access)
vault kv put secret/tmm/development/database \
    host="localhost" \
    port="5433" \
    username="postgres" \
    password="postgres"

vault kv put secret/tmm/development/external \
    stripe_api_key="sk_test_development_key" \
    debug_mode="true"
```

### Staging Environment
```bash
# Staging secrets (production-like but separate)
vault kv put secret/tmm/staging/database \
    host="staging-db.example.com" \
    port="5432" \
    username="tmm_staging" \
    password="staging-secure-password"

vault kv put secret/tmm/staging/external \
    stripe_api_key="sk_test_staging_key" \
    debug_mode="false"
```

### Production Environment
```bash
# Production secrets (highly secure)
vault kv put secret/tmm/production/database \
    host="prod-db.example.com" \
    port="5432" \
    username="tmm_prod" \
    password="highly-secure-production-password"

vault kv put secret/tmm/production/external \
    stripe_api_key="sk_live_production_key" \
    debug_mode="false"
```

## Security Best Practices

### 1. Token Management
```bash
# Create limited-scope tokens
vault write auth/token/create \
    policies="tmm-policy" \
    ttl="1h" \
    renewable=true

# Revoke tokens when done
vault token revoke <token>

# Use AppRole for applications
vault write auth/approle/role/tmm-app \
    token_policies="tmm-policy" \
    token_ttl=30m \
    token_max_ttl=1h
```

### 2. Secret Rotation
```bash
# Rotate database passwords
vault kv put secret/tmm/production/database \
    password="new-rotated-password"

# Update application configuration
kubectl rollout restart deployment/tmm-app
```

### 3. Audit Logging
```bash
# Enable audit logging
vault audit enable file file_path=/vault/logs/audit.log

# View audit logs
tail -f /vault/logs/audit.log | jq .
```

### 4. Backup and Recovery
```bash
# Backup Vault data
vault operator raft snapshot save backup.snap

# Restore from backup
vault operator raft snapshot restore backup.snap
```

## Vault Commands Reference

### Basic Operations
```bash
# Check Vault status
vault status

# List secret engines
vault secrets list

# List authentication methods
vault auth list

# List policies
vault policy list
```

### Secret Operations
```bash
# Create/Update secret
vault kv put secret/path key=value

# Read secret
vault kv get secret/path

# Delete secret
vault kv delete secret/path

# List secrets
vault kv list secret/
```

### Policy Management
```bash
# Create policy
vault policy write policy-name policy.hcl

# Read policy
vault policy read policy-name

# Delete policy
vault policy delete policy-name
```

### Token Operations
```bash
# Create token
vault token create -policy=policy-name

# Lookup token info
vault token lookup

# Renew token
vault token renew

# Revoke token
vault token revoke <token>
```

## Troubleshooting

### Common Issues

**1. Vault sealed**
```bash
# Check seal status
vault status

# Unseal Vault
vault operator unseal <unseal-key>
```

**2. Authentication failed**
```bash
# Check token validity
vault token lookup

# Login with new token
vault auth -method=token token=<new-token>
```

**3. Permission denied**
```bash
# Check current token capabilities
vault token capabilities secret/tmm/development/database

# Update policy if needed
vault policy write tmm-policy - <<EOF
path "secret/data/tmm/*" {
  capabilities = ["read", "list"]
}
EOF
```

**4. Secret not found**
```bash
# List available secrets
vault kv list secret/tmm/development/

# Check secret path format (KV v1 vs v2)
vault kv get secret/tmm/development/database
```

### Development Debugging

#### Enable Debug Logging
```bash
# Set Vault log level
export VAULT_LOG_LEVEL=debug

# Start Vault with debug
vault server -dev -log-level=debug
```

#### Test Secret Access
```bash
# Test secret retrieval
curl -H "X-Vault-Token: dev-root-token" \
     http://localhost:8200/v1/secret/data/tmm/development/database

# Test with application
go run cmd/api/main.go -vault-debug=true
```

## Integration with CI/CD

### GitHub Actions
```yaml
# .github/workflows/deploy.yml
- name: Get secrets from Vault
  uses: hashicorp/vault-action@v2
  with:
    url: ${{ secrets.VAULT_ADDR }}
    token: ${{ secrets.VAULT_TOKEN }}
    secrets: |
      secret/data/tmm/production/database password | DB_PASSWORD
      secret/data/tmm/production/external stripe_api_key | STRIPE_KEY
```

### Docker Deployment
```bash
# Pass Vault credentials to container
docker run -e VAULT_ADDR=$VAULT_ADDR \
           -e VAULT_TOKEN=$VAULT_TOKEN \
           -e VAULT_SECRET_PATH=tmm \
           tmm-app:latest
```

## Advanced Features

### Dynamic Secrets
```bash
# Enable database secrets engine
vault secrets enable database

# Configure database connection
vault write database/config/postgres \
    plugin_name=postgresql-database-plugin \
    connection_url="postgresql://{{username}}:{{password}}@localhost:5432/postgres" \
    allowed_roles="tmm-role"

# Create role for dynamic credentials
vault write database/roles/tmm-role \
    db_name=postgres \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    default_ttl="1h" \
    max_ttl="24h"

# Generate dynamic credentials
vault read database/creds/tmm-role
```

### Secret Templating
```bash
# Use Consul Template for dynamic config
vault kv put secret/tmm/template \
    database_url="postgresql://{{with secret \"secret/tmm/development/database\"}}{{.Data.data.username}}:{{.Data.data.password}}{{end}}@localhost:5432/tmm"
```

**Vault provides enterprise-grade secret management for secure, scalable applications!** 🔐✨ 