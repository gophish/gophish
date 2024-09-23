#!/bin/bash
echo "[*] Building the golang binary into the ./output directory "
docker image build --output "type=local,dest=$(pwd)/output" --target binary .
