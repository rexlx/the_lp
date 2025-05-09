#!/bin/bash

# source .venv/bin/activate
# python3 -m pikepdf --version
# python3 add.py "$@"
# echo "Python script executed successfully."

# Change to the directory where the script is located
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

source .venv/bin/activate
# python3 -m pikepdf --version

# Extract the filename from the path
filename=$(basename "$1")
uuid="$2"

python3 add.py "../static/$filename" "$uuid"
echo "Python script executed successfully."