#!/bin/bash

BIN_DIR=$PWD/bin
OVERRIDES_FILE=$BIN_DIR/developer_overrides.tfrc

cat << EOF > $OVERRIDES_FILE
provider_installation {
  dev_overrides {
    "fastly/fastly" = "$BIN_DIR"
  }
  direct {}
}
EOF

echo ""
echo "A development overrides file has been generated at $OVERRIDES_FILE."
echo "To use your locally built provider, run:"
echo ""
printf "\texport TF_CLI_CONFIG_FILE=$OVERRIDES_FILE\n"
echo ""
