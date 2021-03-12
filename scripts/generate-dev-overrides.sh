#!/bin/bash

BIN_DIR=$PWD/bin

cat << EOF > $BIN_DIR/developer_overrides.tfrc
provider_installation {

  dev_overrides {
    "fastly/fastly" = "$BIN_DIR"
  }

  direct {}
}
EOF

echo ""
echo "A development overrides file has been generated at ./bin/developer_overrides.tfrc."
echo "To make Terraform temporarily use your locally built version of the provider, set TF_CLI_CONFIG_FILE within your terminal"
echo ""
printf "\texport TF_CLI_CONFIG_FILE=$BIN_DIR/developer_overrides.tfrc"
echo ""