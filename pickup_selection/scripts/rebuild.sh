# Stop/remove previous container
./scripts/stop.sh

# Rebuild
docker build --platform linux/amd64 -t pickup-selection .

# Run
./scripts/start_memcached.sh
docker run -d -p 8080:8080 --name pickup-selection --platform linux/amd64 --env-file ./.env --entrypoint /usr/local/bin/aws-lambda-rie pickup-selection ./main
