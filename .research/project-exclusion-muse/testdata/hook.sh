#!/bin/sh
# Record only fixture-local execution and its own process identity.
printf '%s\n' "$$" > "$1.pid"
printf '%s\n' MUSE_EXCL_HOOK > "$1"
