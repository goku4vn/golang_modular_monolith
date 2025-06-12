# Redis with custom configuration
FROM redis:7-alpine

# Copy custom Redis configuration
COPY redis.conf /usr/local/etc/redis/redis.conf

# Set permissions for redis config
RUN chown redis:redis /usr/local/etc/redis/redis.conf

# Switch to redis user
USER redis

# Expose Redis port
EXPOSE 6379

# Run Redis with custom config
CMD ["redis-server", "/usr/local/etc/redis/redis.conf"] 