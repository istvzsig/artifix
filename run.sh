#!/bin/bash
set -euo pipefail

go build -o ./bin/artifix .

exec ./bin/artifix