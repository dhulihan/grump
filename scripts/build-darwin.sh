#!/bin/bash

goreleaser --snapshot --skip-publish --rm-dist -f .goreleaser.yml
