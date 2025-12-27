#!/bin/bash

# Install openapi-typescript if not present
if ! npm list openapi-typescript >/dev/null 2>&1; then
    echo "Installing openapi-typescript..."
    npm install -D openapi-typescript
fi

# Generate types from local spec file
npx openapi-typescript ../api/openapi.yaml -o src/types/api.generated.ts

echo "Types generated successfully!"
