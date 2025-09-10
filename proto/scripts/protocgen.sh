#!/usr/bin/env bash
set -eo pipefail

echo "Generating gogo proto code"

# Run from repo root: enter proto/, generate once using the template,
# then return to root and copy generated files from go_package paths.
cd proto

# IMPORTANT: In proto/buf.gen.gogo.yaml, do NOT set `paths=source_relative`
# for gocosmos. Let go_package decide the output path so it lands under:
#   ../github.com/<module>/...
# Example plugin block:
#   plugins:
#     - name: gocosmos
#       out: ..
#       opt:
#         - plugins=grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types
#         # (no paths=source_relative here)

# Single pass generation for all protos in this module
buf generate --template buf.gen.gogo.yaml

cd ..

# Copy generated code into the repository tree, if present.
# Maany (your modules)
if [ -d github.com/maany-xyz/maany-provider ]; then
  echo "Copying generated files for maany-provider..."
  cp -R github.com/maany-xyz/maany-provider/* ./
fi

# Gaia (metaprotocols module you’re vendoring)
if [ -d github.com/cosmos/gaia ]; then
  echo "Copying generated files for gaia..."
  cp -R github.com/cosmos/gaia/* ./
fi

# Clean the temporary github.com tree created by buf
rm -rf github.com

echo "✅ Protobuf generation complete."
