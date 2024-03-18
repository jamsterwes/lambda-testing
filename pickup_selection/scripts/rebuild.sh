# Stop/remove previous container
./scripts/stop.sh

# Rebuild
docker build --platform linux/arm64 -t pickup-selection .

# Run
docker run -d -p 8080:8080 --name pickup-selection --platform linux/arm64 --env-file ./.env --entrypoint /usr/local/bin/aws-lambda-rie pickup-selection ./main
