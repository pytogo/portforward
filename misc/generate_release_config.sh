#!/bin/bash

# Configuration
BEFORE_SCRIPT_FILE="misc/before_script_linux.sh"

# Generate base CI file
maturin generate-ci github > .github/workflows/release.yml

# Validate input file
if [ ! -f "$BEFORE_SCRIPT_FILE" ]; then
  echo "Error: Input file $BEFORE_SCRIPT_FILE not found"
  exit 1
fi

# Format injection content with proper YAML indentation
INSERT_TEXT="          before-script-linux: |"
while IFS= read -r line; do
  INSERT_TEXT="$INSERT_TEXT\n            $line"
done < "$BEFORE_SCRIPT_FILE"

# Create temporary file for processing
TMP_FILE=$(mktemp)

# Process file line-by-line and inject content after each "manylinux:" match
while IFS= read -r line; do
  echo "$line" >> "$TMP_FILE"
  
  if echo "$line" | grep -q "manylinux:"; then
    echo -e "$INSERT_TEXT" >> "$TMP_FILE"
  fi
done < .github/workflows/release.yml

# Replace original file with processed version
mv "$TMP_FILE" .github/workflows/release.yml

# Update workflow name - handle BSD vs GNU sed differently
if sed --version 2>/dev/null | grep -q GNU; then
  # GNU sed
  sed -i -e 's/^name: CI$/name: Release/' .github/workflows/release.yml
else
  # BSD sed (macOS)
  sed -i '' -e 's/^name: CI$/name: Release/' .github/workflows/release.yml
fi

echo "Done! CI workflow updated with external content."
