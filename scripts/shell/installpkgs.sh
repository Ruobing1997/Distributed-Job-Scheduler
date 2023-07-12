#!/bin/bash

# Check if requests is installed
if ! pip freeze | grep -q 'requests=='; then
    echo "Installing requests library..."
    pip install requests
else
    echo "requests library is already installed."
fi