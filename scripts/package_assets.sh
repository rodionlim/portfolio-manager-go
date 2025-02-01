#!/usr/bin/env bash
#
# compress static assets

set -euo pipefail

version="$(< VERSION)"
mkdir -p .tarballs
cd web/ui
find static -type f -not -name '*.gz' -print0 | xargs -0 tar czf ../../.tarballs/portfolio-manager-web-ui-${version}.tar.gz
