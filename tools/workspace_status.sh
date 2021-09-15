#!/bin/bash

echo "STABLE_GIT_SHA1 $(git rev-parse HEAD)"
echo "STABLE_GIT_SHA1_SHORT $(git rev-parse HEAD | head -c 10)"

[[ -z "${GITHUB_REF}" ]] && echo "STABLE_GIT_BRANCH $(git rev-parse --abbrev-ref HEAD)" || echo "STABLE_GIT_BRANCH ${GITHUB_REF#refs/heads/}"
