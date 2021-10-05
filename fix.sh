#!/bin/bash
find . -type f -name '*.go' -exec sed -i '' -e 's#github.com/owncast/owncast#github.com/dolfly/owncast#g' {} \;
