#!/usr/bin/env bash
set -eo pipefail

echo "Generating gogo proto code"

cd proto

buf generate --template buf.gen.gogo.yaml

cd ..

if [ -d github.com/maany-xyz/maany-provider ]; then
  echo "Copying generated files for maany-provider..."
  cp -R github.com/maany-xyz/maany-provider/* ./
fi

rm -rf github.com

echo "✅ Protobuf generation complete."
