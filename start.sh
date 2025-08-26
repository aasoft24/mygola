#!/bin/bash
echo "Starting MyGola with Watchexec watcher..."
watchexec -r -e go,html,tpl,yaml "go run main.go serve"
