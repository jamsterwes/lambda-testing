# Kill container if running
docker kill pickup-selection || true

# Kill memcached if running
docker kill memcached || true

# Remove container if exists
docker rm pickup-selection || true