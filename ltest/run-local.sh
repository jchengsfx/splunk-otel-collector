#!/usr/bin/env bash

export SFX_RUN_MONITORS=1
export SIGNALFX_BUNDLE_DIR=/home/rmfitzpatrick/signalfx-agent
./bin/otelcol --config=./ltest/sfx_soc_config.yaml

