#!/usr/bin/env bash
go run . ./testdata/hello/hello >&2 out && cat out
