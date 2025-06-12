# HashiCorp Vault with custom configuration
FROM hashicorp/vault:latest

# Copy configuration files
COPY vault-config.hcl /vault/config/vault-config.hcl
COPY vault-init.sh /vault/scripts/vault-init.sh
COPY vault.env /vault/config/vault.env

# Make init script executable
RUN chmod +x /vault/scripts/vault-init.sh

# Create vault user and set permissions
USER vault

# Expose Vault port
EXPOSE 8200

# Set environment variables
ENV VAULT_ADDR=http://0.0.0.0:8200
ENV VAULT_CONFIG_PATH=/vault/config/vault-config.hcl

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD vault status || exit 1

# Run Vault server
CMD ["vault", "server", "-config=/vault/config/vault-config.hcl"] 