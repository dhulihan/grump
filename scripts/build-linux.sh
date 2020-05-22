#!/bin/bash

goreleaser --skip-publish --rm-dist -f .goreleaser-linux.yml
