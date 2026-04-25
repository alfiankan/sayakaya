#!/bin/bash

./run_dev.sh

./test_client.sh

docker-compose down -v

echo "Verification complete!"