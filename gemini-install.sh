#!/usr/bin/env bash

gemini extensions uninstall cicd

./build.sh
mkdir -p ~/.gemini/extensions/cicd/bin
cp cicd-mcp-server/cicd-mcp-server ~/.gemini/extensions/cicd/bin/
cp gemini-extension.json ~/.gemini/extensions/cicd
cp -r skills ~/.gemini/extensions/cicd/
