#!/usr/bin/env bash

# Test script for Compute service orchestration lifecycle
# Tests: fastly_service_compute, fastly_service_domain, fastly_service_backend,
#        fastly_compute_package_upload, fastly_service_version_clone,
#        and fastly_service_version_activate actions

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$REPO_ROOT/test-orchestration-compute-$$"
TF_CONFIG_DIR="$SCRIPT_DIR"
PACKAGE_PATH="$REPO_ROOT/internal/acceptance_tests/fixtures/packages/valid.tar.gz"

# Test configuration
TEST_SERVICE_1_NAME="tf-test-compute-svc1-$$"
TEST_SERVICE_2_NAME="tf-test-compute-svc2-$$"
SERVICE_1_ID=""
SERVICE_2_ID=""

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_step() {
    echo -e "\n${BLUE}===${NC} $* ${BLUE}===${NC}\n"
}

# Cleanup function (for emergency cleanup on failures)
cleanup() {
    local exit_code=$?

    if [ $exit_code -ne 0 ]; then
        log_error "Script failed with exit code $exit_code"
        log_step "Emergency cleanup"

        # Only do emergency cleanup if the test didn't complete normally
        if [ -n "$SERVICE_1_ID" ] || [ -n "$SERVICE_2_ID" ]; then
            log_info "Attempting emergency cleanup of services..."

            # Deactivate and delete services
            for version in 1 2 3 4 5; do
                [ -n "$SERVICE_1_ID" ] && curl -s -X PUT -H "Fastly-Key: $FASTLY_API_TOKEN" \
                    "https://api.fastly.com/service/$SERVICE_1_ID/version/$version/deactivate" > /dev/null 2>&1 || true
                [ -n "$SERVICE_2_ID" ] && curl -s -X PUT -H "Fastly-Key: $FASTLY_API_TOKEN" \
                    "https://api.fastly.com/service/$SERVICE_2_ID/version/$version/deactivate" > /dev/null 2>&1 || true
            done

            [ -n "$SERVICE_1_ID" ] && curl -s -X DELETE -H "Fastly-Key: $FASTLY_API_TOKEN" \
                "https://api.fastly.com/service/$SERVICE_1_ID" > /dev/null 2>&1 || true
            [ -n "$SERVICE_2_ID" ] && curl -s -X DELETE -H "Fastly-Key: $FASTLY_API_TOKEN" \
                "https://api.fastly.com/service/$SERVICE_2_ID" > /dev/null 2>&1 || true

            log_info "Emergency cleanup completed"
        fi
    fi

    # Remove test directory
    if [ -d "$TEST_DIR" ]; then
        cd "$REPO_ROOT"
        rm -rf "$TEST_DIR"
        log_info "Removed test directory: $TEST_DIR"
    fi

    exit $exit_code
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites"

    # Check for required environment variables
    if [ -z "${FASTLY_API_TOKEN:-}" ]; then
        log_error "FASTLY_API_TOKEN environment variable is not set"
        exit 1
    fi
    log_success "FASTLY_API_TOKEN is set"

    # Check for required commands
    local required_commands=("terraform" "go" "jq")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "Required command '$cmd' not found"
            exit 1
        fi
        log_success "Found command: $cmd"
    done

    # Check for package file
    if [ ! -f "$PACKAGE_PATH" ]; then
        log_error "Compute package not found at: $PACKAGE_PATH"
        exit 1
    fi
    log_success "Found compute package: $PACKAGE_PATH"
}

# Build the provider
build_provider() {
    log_step "Building Terraform provider"

    cd "$REPO_ROOT"

    log_info "Running go build..."
    go build -o terraform-provider-fastly

    local provider_path="$REPO_ROOT/terraform-provider-fastly"
    if [ ! -f "$provider_path" ]; then
        log_error "Provider binary not found at $provider_path"
        exit 1
    fi

    log_success "Provider built successfully: $provider_path"

    # Set up Terraform to use local provider
    export TF_CLI_CONFIG_FILE="$TEST_DIR/.terraformrc"
}

# Create test directory and configuration
setup_test_environment() {
    log_step "Setting up test environment"

    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    log_info "Copying Terraform configuration files..."

    # Copy Terraform configuration files from the config directory
    cp "$TF_CONFIG_DIR/main.tf" .
    cp "$TF_CONFIG_DIR/variables.tf" .
    cp "$TF_CONFIG_DIR/outputs.tf" .

    # Create terraform.tfvars
    cat > terraform.tfvars << EOF
fastly_api_token  = "$FASTLY_API_TOKEN"
service_1_name    = "$TEST_SERVICE_1_NAME"
service_1_version = 1
service_1_domain  = "test-compute-svc1-$$.example.com"
service_2_name    = "$TEST_SERVICE_2_NAME"
service_2_version = 1
service_2_domain  = "test-compute-svc2-$$.example.com"
package_path      = "$PACKAGE_PATH"
EOF

    # Create .terraformrc for local provider
    cat > .terraformrc << EOF
provider_installation {
  dev_overrides {
    "fastly/fastly" = "$REPO_ROOT"
  }
  direct {}
}
EOF

    log_success "Test environment created at: $TEST_DIR"
}

# Initialize and validate Terraform
init_terraform() {
    log_step "Initializing Terraform"

    cd "$TEST_DIR"

    log_info "Running terraform init..."
    terraform init

    log_info "Running terraform validate..."
    terraform validate

    log_success "Terraform initialized and validated"
}

# Apply initial configuration
apply_initial_config() {
    log_step "Applying initial Terraform configuration"

    cd "$TEST_DIR"

    log_info "Running terraform plan..."
    terraform plan -out=tfplan

    log_info "Running terraform apply..."
    terraform apply tfplan

    # Extract service IDs
    SERVICE_1_ID=$(terraform output -raw service_1_id)
    SERVICE_2_ID=$(terraform output -raw service_2_id)

    log_success "Initial configuration applied"
    log_info "Service 1 ID: $SERVICE_1_ID"
    log_info "Service 2 ID: $SERVICE_2_ID"
}

# Verify initial state
verify_initial_state() {
    log_step "Verifying initial state"

    cd "$TEST_DIR"

    local svc1_active=$(terraform output -raw service_1_active_version)
    local svc1_latest=$(terraform output -raw service_1_latest_version)
    local svc2_active=$(terraform output -raw service_2_active_version)
    local svc2_latest=$(terraform output -raw service_2_latest_version)

    log_info "Service 1 - Active: $svc1_active, Latest: $svc1_latest"
    log_info "Service 2 - Active: $svc2_active, Latest: $svc2_latest"

    # Verify resources exist
    if [ -z "$svc1_active" ] || [ "$svc1_active" = "0" ]; then
        log_error "Service 1 has no active version"
        exit 1
    fi

    if [ -z "$svc2_active" ] || [ "$svc2_active" = "0" ]; then
        log_error "Service 2 has no active version"
        exit 1
    fi

    log_success "Initial state verified"
}

# Test compute package upload action
test_package_upload_action() {
    log_step "Testing compute package upload action"

    cd "$TEST_DIR"

    log_info "Invoking package upload action for service 1..."

    if terraform apply -invoke=action.fastly_compute_package_upload.service_1_upload -auto-approve; then
        log_success "Package upload action invoked successfully for service 1"
    else
        log_error "Failed to invoke package upload action for service 1"
        return 1
    fi

    log_info "Invoking package upload action for service 2..."

    if terraform apply -invoke=action.fastly_compute_package_upload.service_2_upload -auto-approve; then
        log_success "Package upload action invoked successfully for service 2"
    else
        log_error "Failed to invoke package upload action for service 2"
        return 1
    fi

    log_success "Package upload actions completed"
}

# Test version cloning with actions
test_version_clone_action() {
    log_step "Testing version clone action"

    cd "$TEST_DIR"

    local svc1_active=$(terraform output -raw service_1_active_version)
    local svc1_latest=$(terraform output -raw service_1_latest_version)
    log_info "Service 1 - Active: $svc1_active, Latest: $svc1_latest"

    # Check Terraform version
    local tf_version=$(terraform version -json | jq -r '.terraform_version')
    log_info "Terraform version: $tf_version"

    # Invoke the clone action for service 1
    log_info "Invoking version clone action for service 1 (cloning version $svc1_active)..."

    if terraform apply -invoke=action.fastly_service_version_clone.service_1_clone -auto-approve; then
        log_success "Version clone action invoked successfully for service 1"

        # Refresh data to see the new version
        terraform refresh > /dev/null
        local new_latest=$(terraform output -raw service_1_latest_version)
        log_info "Service 1 new latest version: $new_latest (was $svc1_latest)"

        if [ "$new_latest" -gt "$svc1_latest" ]; then
            log_success "Version cloned successfully - version incremented from $svc1_latest to $new_latest"
        else
            log_warning "Latest version did not increment (still $new_latest)"
        fi
    else
        log_error "Failed to invoke version clone action for service 1"
        return 1
    fi

    # Test cloning for service 2
    local svc2_active=$(terraform output -raw service_2_active_version)
    local svc2_latest=$(terraform output -raw service_2_latest_version)
    log_info "Service 2 - Active: $svc2_active, Latest: $svc2_latest"

    log_info "Invoking version clone action for service 2 (cloning version $svc2_active)..."

    if terraform apply -invoke=action.fastly_service_version_clone.service_2_clone -auto-approve; then
        log_success "Version clone action invoked successfully for service 2"

        terraform refresh > /dev/null
        local new_latest=$(terraform output -raw service_2_latest_version)
        log_info "Service 2 new latest version: $new_latest (was $svc2_latest)"

        if [ "$new_latest" -gt "$svc2_latest" ]; then
            log_success "Version cloned successfully - version incremented from $svc2_latest to $new_latest"
        else
            log_warning "Latest version did not increment (still $new_latest)"
        fi
    else
        log_error "Failed to invoke version clone action for service 2"
        return 1
    fi

    log_success "Version clone actions completed"
}

# Test version activation with actions
test_version_activate_action() {
    log_step "Testing version activate action"

    cd "$TEST_DIR"

    # Update service_1_version to 2 (the cloned version)
    log_info "Updating terraform.tfvars to set service versions to 2..."
    cat > terraform.tfvars << EOF
fastly_api_token  = "$FASTLY_API_TOKEN"
service_1_name    = "$TEST_SERVICE_1_NAME"
service_1_version = 2
service_1_domain  = "test-compute-svc1-$$.example.com"
service_2_name    = "$TEST_SERVICE_2_NAME"
service_2_version = 2
service_2_domain  = "test-compute-svc2-$$.example.com"
package_path      = "$PACKAGE_PATH"
EOF

    # Need to upload package to version 2 before activating
    log_info "Uploading package to service 1 version 2..."
    if terraform apply -invoke=action.fastly_compute_package_upload.service_1_upload -auto-approve; then
        log_success "Package uploaded to service 1 version 2"
    else
        log_error "Failed to upload package to service 1 version 2"
        return 1
    fi

    local svc1_active_before=$(terraform output -raw service_1_active_version)
    log_info "Service 1 active version before activation: $svc1_active_before"

    log_info "Invoking version activate action for service 1 version 2..."

    if terraform apply -invoke=action.fastly_service_version_activate.service_1_activate -auto-approve; then
        log_success "Version activate action invoked successfully for service 1"

        # Refresh to get updated active version
        terraform refresh > /dev/null
        local svc1_active_after=$(terraform output -raw service_1_active_version)
        log_info "Service 1 active version after activation: $svc1_active_after"

        if [ "$svc1_active_after" = "2" ]; then
            log_success "Version activated successfully - active version is now 2"
        else
            log_error "Active version is $svc1_active_after, expected 2"
            return 1
        fi
    else
        log_error "Failed to invoke version activate action for service 1"
        return 1
    fi

    log_success "Version activate action completed"
}

# Test resource updates
test_resource_updates() {
    log_step "Testing resource updates"

    cd "$TEST_DIR"

    log_info "Updating service 1 comment..."

    # Update the service comment in the config
    sed -i.bak 's/Test compute service 1 for orchestration/Test compute service 1 for orchestration - UPDATED/' main.tf

    log_info "Running terraform plan..."
    terraform plan -out=tfplan

    log_info "Running terraform apply..."
    terraform apply tfplan

    log_success "Service update completed"
}

# Test resource destruction
test_resource_destruction() {
    log_step "Testing resource destruction"

    cd "$TEST_DIR"

    # Remove version-locked resources from state
    log_info "Removing version-locked resources from state..."
    terraform state rm fastly_service_domain.service_1_domain || true
    terraform state rm fastly_service_backend.service_1_backend_shared || true
    terraform state rm fastly_service_backend.service_1_backend_unique || true
    terraform state rm fastly_service_domain.service_2_domain || true
    terraform state rm fastly_service_backend.service_2_backend_shared || true

    log_info "Running terraform destroy..."
    terraform destroy -auto-approve

    log_success "Resources destroyed"

    # Verify services were deleted
    log_info "Verifying services were deleted..."

    local svc1_exists=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_1_ID" | jq -r '.id // empty')

    if [ -z "$svc1_exists" ]; then
        log_success "Service 1 successfully deleted"
    else
        log_error "Service 1 still exists"
        return 1
    fi

    local svc2_exists=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_2_ID" | jq -r '.id // empty')

    if [ -z "$svc2_exists" ]; then
        log_success "Service 2 successfully deleted"
    else
        log_error "Service 2 still exists"
        return 1
    fi

    # Clear service IDs to prevent emergency cleanup
    SERVICE_1_ID=""
    SERVICE_2_ID=""

    log_success "Service destruction verified"
}

# Main test execution
main() {
    log_step "Starting Compute Service Orchestration Lifecycle Tests"

    check_prerequisites
    build_provider
    setup_test_environment
    init_terraform
    apply_initial_config
    verify_initial_state
    test_package_upload_action
    test_version_clone_action
    test_version_activate_action
    test_resource_updates
    test_resource_destruction

    log_step "All tests passed!"
    log_success "Compute service orchestration lifecycle tests completed successfully"
}

# Run main
main
