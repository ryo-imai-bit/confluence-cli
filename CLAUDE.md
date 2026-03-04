# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI tool for Atlassian Confluence, built with Go for portability.

## API Reference

- Uses Confluence REST API v2
- OpenAPI spec: `openapi-v2.v3.json`
- API docs: https://developer.atlassian.com/cloud/confluence/rest/v2/intro#about

## Development

This is an early-stage project. The codebase will be implemented in Go.

### Build Commands (once Go code exists)

```bash
go build -o confluence-cli .
go test ./...
go test -v ./... -run TestName  # run single test
```

## Architecture Notes

- CLI interface design should be informed by available Confluence API capabilities
- Focus on portability for distribution as a standalone binary
