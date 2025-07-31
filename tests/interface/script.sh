#!/usr/bin/env bash

cd ./tests/interface/ || exit

echo DEPLOYING USING LATEST TERRAFORM VERSION
terraform init -upgrade
terraform apply -auto-approve

# cleanup() {
# 	# reset back to the installed provider so we can destroy the service
# 	unset TF_CLI_CONFIG_FILE
	# terraform init
	# echo ""
	# echo "Running terraform destroy..."
	# terraform destroy -auto-approve
# }
# trap cleanup EXIT

cd - || exit
make build
BIN_DIR=$PWD/bin
OVERRIDES_FILENAME=developer_overrides.tfrc
export TF_CLI_CONFIG_FILE="$BIN_DIR/$OVERRIDES_FILENAME"
unset TF_CLI_CONFIG_FILE
terraform init
echo ""
echo "Running terraform destroy..."
terraform destroy -auto-approve
# cd - || exit

# echo RUNNING PLAN USING TERRAFORM VERSION BUILT FROM THIS BRANCH
# plan_output=$(terraform plan -no-color 2>&1)

# if [[ "$plan_output" == *"No changes. Your infrastructure matches the configuration."* ]]; then
# 	echo ""
# 	echo "Terraform plan succeeded: No changes detected."
# 	exit 0
# else
# 	echo ""
# 	echo "Terraform plan failed: Changes detected or unexpected output."
# 	echo "$plan_output"
# 	exit 1
# fi
