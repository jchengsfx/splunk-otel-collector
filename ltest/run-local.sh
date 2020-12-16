#!/usr/bin/env bash

export SIGNALFX_BUNDLE_DIR=/home/rmfitzpatrick/signalfx-agent
./bin/otelcol --config=./ltest/sfx_soc_config.yaml

