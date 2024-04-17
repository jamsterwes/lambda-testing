# Kill container if running
docker kill pickup-selection || true

# Kill memcached if running
docker kill prom-memcached || true

# Remove container if exists
docker rm pickup-selection || true

# Remove memcached if exists
docker rm prom-memcached || true