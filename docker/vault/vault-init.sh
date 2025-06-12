#!/bin/bash

# Wait for Vault to be ready
echo "🔒 Waiting for Vault to be ready..."
until curl -s http://localhost:8200/v1/sys/health > /dev/null 2>&1; do
    sleep 2
done

echo "🔒 Vault is ready, initializing..."

# Initialize Vault (development mode)
export VAULT_ADDR='http://localhost:8200'

# Check if Vault is already initialized
if vault status | grep -q "Initialized.*true"; then
    echo "🔒 Vault is already initialized"
    exit 0
fi

# Initialize Vault
echo "🔒 Initializing Vault..."
vault operator init -key-shares=1 -key-threshold=1 -format=json > /tmp/vault-keys.json

# Extract unseal key and root token
UNSEAL_KEY=$(cat /tmp/vault-keys.json | jq -r '.unseal_keys_b64[0]')
ROOT_TOKEN=$(cat /tmp/vault-keys.json | jq -r '.root_token')

# Unseal Vault
echo "🔓 Unsealing Vault..."
vault operator unseal $UNSEAL_KEY

# Login with root token
echo "🔑 Logging in with root token..."
vault auth $ROOT_TOKEN

# Enable KV secrets engine v2
echo "🗄️ Enabling KV secrets engine..."
vault secrets enable -version=2 kv

# Create sample secrets for development (organized by module)
echo "📝 Creating sample secrets organized by module..."

# App-level secrets
echo "📱 Creating app secrets..."
vault kv put kv/app \
    APP_VERSION="2.1.0" \
    APP_NAME="modular-monolith-vault" \
    GIN_MODE="release" \
    PORT="8080"

# Customer module secrets
echo "👤 Creating customer module secrets..."
vault kv put kv/modules/customer \
    DATABASE_HOST="postgres" \
    DATABASE_PORT="5432" \
    DATABASE_USER="postgres" \
    DATABASE_PASSWORD="vault_customer_password" \
    DATABASE_NAME="modular_monolith_customer" \
    DATABASE_SSLMODE="disable" \
    API_KEY="customer_api_key_secret" \
    ENCRYPTION_KEY="customer_encryption_key"

# Order module secrets
echo "📦 Creating order module secrets..."
vault kv put kv/modules/order \
    DATABASE_HOST="postgres" \
    DATABASE_PORT="5432" \
    DATABASE_USER="postgres" \
    DATABASE_PASSWORD="vault_order_password" \
    DATABASE_NAME="modular_monolith_order" \
    DATABASE_SSLMODE="disable" \
    PAYMENT_API_KEY="order_payment_api_secret" \
    WEBHOOK_SECRET="order_webhook_secret"

# Product module secrets
echo "🛍️ Creating product module secrets..."
vault kv put kv/modules/product \
    DATABASE_HOST="postgres" \
    DATABASE_PORT="5432" \
    DATABASE_USER="postgres" \
    DATABASE_PASSWORD="vault_product_password" \
    DATABASE_NAME="modular_monolith_product" \
    DATABASE_SSLMODE="disable" \
    INVENTORY_API_KEY="product_inventory_api_secret" \
    CACHE_KEY="product_cache_secret"

# Create AppRole for application authentication
echo "🔐 Setting up AppRole authentication..."
vault auth enable approle

# Create policy for the application (access to all modules)
vault policy write modular-monolith-policy - <<EOF
# App-level secrets
path "kv/data/app" {
  capabilities = ["read"]
}
path "kv/metadata/app" {
  capabilities = ["read"]
}

# Customer module secrets
path "kv/data/modules/customer" {
  capabilities = ["read"]
}
path "kv/metadata/modules/customer" {
  capabilities = ["read"]
}

# Order module secrets
path "kv/data/modules/order" {
  capabilities = ["read"]
}
path "kv/metadata/modules/order" {
  capabilities = ["read"]
}

# Product module secrets
path "kv/data/modules/product" {
  capabilities = ["read"]
}
path "kv/metadata/modules/product" {
  capabilities = ["read"]
}
EOF

# Create AppRole
vault write auth/approle/role/modular-monolith \
    token_policies="modular-monolith-policy" \
    token_ttl=1h \
    token_max_ttl=4h

# Get RoleID and SecretID
ROLE_ID=$(vault read -field=role_id auth/approle/role/modular-monolith/role-id)
SECRET_ID=$(vault write -field=secret_id -f auth/approle/role/modular-monolith/secret-id)

echo "🎉 Vault initialization completed!"
echo "📋 Vault Details:"
echo "   - Vault Address: http://localhost:8200"
echo "   - Root Token: $ROOT_TOKEN"
echo "   - Unseal Key: $UNSEAL_KEY"
echo "   - Role ID: $ROLE_ID"
echo "   - Secret ID: $SECRET_ID"
echo ""
echo "🔧 Environment Variables for Application:"
echo "   export VAULT_ENABLED=true"
echo "   export VAULT_ADDR=http://localhost:8200"
echo "   export VAULT_ROLE_ID=$ROLE_ID"
echo "   export VAULT_SECRET_ID=$SECRET_ID"
echo "   export VAULT_MOUNT_PATH=kv"
echo "   export VAULT_SECRET_PATH=modular-monolith"
echo ""
echo "🌐 Vault UI: http://localhost:8200/ui"
echo "   Login with token: $ROOT_TOKEN"

# Save credentials for later use
cat > /vault/credentials.txt <<EOF
ROOT_TOKEN=$ROOT_TOKEN
UNSEAL_KEY=$UNSEAL_KEY
ROLE_ID=$ROLE_ID
SECRET_ID=$SECRET_ID
EOF

echo "💾 Credentials saved to /vault/credentials.txt" 