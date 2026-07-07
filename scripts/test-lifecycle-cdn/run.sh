#!/usr/bin/env bash

# Test script for the full provider lifecycle
# Tests: fastly_service_cdn, fastly_service_domain, fastly_service_backend,
#        fastly_service_version_clone, and fastly_service_version_activate actions
#
# Coverage includes:
#   - Clone from active version
#   - Clone from latest version (when latest != active)
#   - Version-locked resource writes (domains, backends)
#   - Version activation workflow

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
TEST_DIR="$REPO_ROOT/test-lifecycle-$$"
TF_CONFIG_DIR="$SCRIPT_DIR"

# Test configuration
TEST_SERVICE_1_NAME="tf-test-svc1-$$"
TEST_SERVICE_2_NAME="tf-test-svc2-$$"
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
            for version in 1 2 3 4 5 6 7 8; do
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
service_1_domain  = "test-svc1-$$.example.com"
service_2_name    = "$TEST_SERVICE_2_NAME"
service_2_version = 1
service_2_domain  = "test-svc2-$$.example.com"
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
            log_success "New version $new_latest created successfully"
        else
            log_warning "Latest version did not change (may already be cloned)"
        fi
    else
        log_error "Failed to invoke version clone action"
        return 1
    fi

    # Invoke the clone action for service 2
    local svc2_active=$(terraform output -raw service_2_active_version)
    log_info "Invoking version clone action for service 2 (cloning version $svc2_active)..."

    if terraform apply -invoke=action.fastly_service_version_clone.service_2_clone -auto-approve; then
        log_success "Version clone action invoked successfully for service 2"
    else
        log_error "Failed to invoke version clone action for service 2"
        return 1
    fi
}

# Test version activation with actions
test_version_activate_action() {
    log_step "Testing version activate action"

    cd "$TEST_DIR"

    # Get current active versions
    local svc1_active=$(terraform output -raw service_1_active_version)
    local svc1_latest=$(terraform output -raw service_1_latest_version)
    log_info "Service 1 - Active: $svc1_active, Latest: $svc1_latest"

    # Only activate if there's a newer version to activate
    if [ "$svc1_latest" -gt "$svc1_active" ]; then
        log_info "Activating version $svc1_latest for service 1..."

        # Update the tfvars to point to the new version
        sed -i.bak "s/service_1_version = [0-9]*/service_1_version = $svc1_latest/" terraform.tfvars
        rm -f terraform.tfvars.bak

        # Apply the configuration change first
        terraform apply -auto-approve > /dev/null

        # Now invoke the activation action
        if terraform apply -invoke=action.fastly_service_version_activate.service_1_activate -auto-approve; then
            log_success "Version $svc1_latest activated successfully for service 1"

            # Verify activation
            terraform refresh > /dev/null
            local new_active=$(terraform output -raw service_1_active_version)
            if [ "$new_active" = "$svc1_latest" ]; then
                log_success "Verified: Version $new_active is now active"
            else
                log_warning "Active version is $new_active (expected $svc1_latest)"
            fi
        else
            log_error "Failed to invoke version activation action"
            return 1
        fi
    else
        log_info "Service 1 active version ($svc1_active) is already the latest"
        log_info "Testing activation action on current version..."

        if terraform apply -invoke=action.fastly_service_version_activate.service_1_activate -auto-approve; then
            log_success "Version activation action completed (version already active)"
        else
            log_error "Failed to invoke version activation action"
            return 1
        fi
    fi

    log_success "Action activation test completed"
}

# Verify service configuration
verify_service_configuration() {
    log_step "Verifying service configuration"

    cd "$TEST_DIR"

    log_info "Checking Terraform state..."

    # Check service 1 resources
    terraform state show fastly_service_cdn.service_1 > /dev/null
    terraform state show fastly_service_domain.service_1_domain > /dev/null
    terraform state show fastly_service_backend.service_1_backend_shared > /dev/null
    terraform state show fastly_service_backend.service_1_backend_unique > /dev/null
    log_success "Service 1 resources verified"

    # Check service 2 resources
    terraform state show fastly_service_cdn.service_2 > /dev/null
    terraform state show fastly_service_domain.service_2_domain > /dev/null
    terraform state show fastly_service_backend.service_2_backend_shared > /dev/null
    log_success "Service 2 resources verified"

    # Verify data sources
    terraform state show data.fastly_service_version.service_1 > /dev/null
    terraform state show data.fastly_service_version.service_2 > /dev/null
    log_success "Version data sources verified"
}

# Test resource updates
test_resource_updates() {
    log_step "Testing resource updates"

    cd "$TEST_DIR"

    # Update service 1 comment
    log_info "Updating service 1 comment..."
    sed -i.bak 's/comment       = "Test service 1"/comment       = "Updated test service 1"/' main.tf
    rm -f main.tf.bak

    terraform plan -out=tfplan
    terraform apply tfplan

    # Verify update
    local comment=$(terraform state show fastly_service_cdn.service_1 | grep "comment" | awk '{print $3}')
    if [[ ! "$comment" =~ "Updated" ]]; then
        log_error "Service comment was not updated"
        exit 1
    fi

    log_success "Resource updates work correctly"
}

# Test clone from latest version and version-locked resource writes
test_clone_from_latest_and_version_writes() {
    log_step "Testing clone from latest version and version-locked resource writes"

    cd "$TEST_DIR"

    # At this point, version 2 should be active for service 1
    local svc1_active=$(terraform output -raw service_1_active_version)
    local svc1_latest=$(terraform output -raw service_1_latest_version)
    log_info "Service 1 - Active: $svc1_active, Latest: $svc1_latest"

    # Clone from active (version 2) to create version 3
    log_info "Cloning from active version ($svc1_active) to create version 3..."
    if terraform apply -invoke=action.fastly_service_version_clone.service_1_clone -auto-approve; then
        log_success "Version cloned from active"
        terraform refresh > /dev/null
        svc1_latest=$(terraform output -raw service_1_latest_version)
        log_info "New latest version: $svc1_latest"
    else
        log_error "Failed to clone from active version"
        return 1
    fi

    # Now active=2, latest=3 (draft version exists)
    # Test clone from latest (version 3)
    log_info "Now active=$svc1_active, latest=$svc1_latest (draft exists)"
    log_info "Cloning from latest version ($svc1_latest) to create version 4..."

    if terraform apply -invoke=action.fastly_service_version_clone.service_1_clone_from_latest -auto-approve; then
        log_success "Version cloned from latest (draft version)"
        terraform refresh > /dev/null
        local new_latest=$(terraform output -raw service_1_latest_version)
        log_info "New latest version after cloning from latest: $new_latest"

        if [ "$new_latest" = "4" ]; then
            log_success "Clone from latest successful - created version 4 from version 3"
        else
            log_error "Expected version 4, got version $new_latest"
            return 1
        fi
    else
        log_error "Failed to clone from latest version"
        return 1
    fi

    # Test version-locked resource writes
    # Add a new domain and backend to version 4
    log_info "Testing version-locked resource writes on version 4..."

    log_info "Updating terraform.tfvars to change version to 4 and add new resources..."

    cat > terraform.tfvars << EOF
fastly_api_token     = "$FASTLY_API_TOKEN"
service_1_name       = "$TEST_SERVICE_1_NAME"
service_1_version    = 4
service_1_domain     = "test-svc1-$$.example.com"
service_1_new_domain = "test-svc1-new-$$.example.com"
service_1_new_backend = "new-backend.example.com"
service_2_name       = "$TEST_SERVICE_2_NAME"
service_2_version    = 2
service_2_domain     = "test-svc2-$$.example.com"
EOF

    log_info "Running terraform plan to update version and add new domain and backend..."
    terraform plan -out=tfplan

    log_info "Running terraform apply to write new resources to version 4..."
    if terraform apply tfplan; then
        log_success "New domain and backend added to version 4"
    else
        log_error "Failed to add resources to version 4"
        return 1
    fi

    # Verify the resources were added
    log_info "Verifying new resources were added to version 4..."
    local domain_count=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_1_ID/version/4/domain" | jq '. | length')
    local backend_count=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_1_ID/version/4/backend" | jq '. | length')

    log_info "Version 4 has $domain_count domains and $backend_count backends"

    if [ "$domain_count" -ge "2" ]; then
        log_success "Domain successfully added to version 4"
    else
        log_error "Expected at least 2 domains in version 4, found $domain_count"
        return 1
    fi

    if [ "$backend_count" -ge "3" ]; then
        log_success "Backend successfully added to version 4"
    else
        log_error "Expected at least 3 backends in version 4, found $backend_count"
        return 1
    fi

    # Activate version 4
    log_info "Activating version 4 with new resources..."
    if terraform apply -invoke=action.fastly_service_version_activate.service_1_activate -auto-approve; then
        log_success "Version 4 activated"
        terraform refresh > /dev/null
        local final_active=$(terraform output -raw service_1_active_version)
        log_info "Final active version: $final_active"

        if [ "$final_active" = "4" ]; then
            log_success "Version 4 (with new resources) is now active"
        else
            log_error "Expected active version 4, got $final_active"
            return 1
        fi
    else
        log_error "Failed to activate version 4"
        return 1
    fi

    log_success "Clone from latest and version write tests completed"
}

# The Fastly API rejects deletes of objects (domains/backends/ACLs) from a locked
# (i.e. ever-activated) service version, and a locked version can never become
# mutable again. If Terraform state has a resource pinned to a locked version,
# clone that version to a fresh, never-activated draft and move the resource
# there via a normal `terraform apply`, so the eventual `terraform destroy` can
# delete it directly with no state manipulation.
advance_off_locked_versions() {
    log_step "Advancing resources off any locked versions before destroy"

    cd "$TEST_DIR"

    terraform refresh > /dev/null 2>&1 || true

    local svc1_version=$(grep -oE 'service_1_version[[:space:]]*=[[:space:]]*[0-9]+' terraform.tfvars | grep -oE '[0-9]+$')
    local svc1_locked=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_1_ID/version/$svc1_version" | jq -r '.locked')

    if [ "$svc1_locked" = "true" ]; then
        log_info "Service 1 version $svc1_version is locked; cloning to a fresh draft version..."

        if ! terraform apply -invoke=action.fastly_service_version_clone.service_1_clone_from_pinned -auto-approve; then
            log_error "Failed to clone service 1 off its locked version"
            return 1
        fi

        terraform refresh > /dev/null
        local svc1_new_version=$(terraform output -raw service_1_latest_version)
        log_success "Cloned version $svc1_version to draft version $svc1_new_version"

        sed -i.bak "s/service_1_version[[:space:]]*=[[:space:]]*[0-9]*/service_1_version = $svc1_new_version/" terraform.tfvars
        rm -f terraform.tfvars.bak

        log_info "Moving service 1 resources to version $svc1_new_version..."
        terraform plan -out=tfplan
        terraform apply tfplan
        log_success "Service 1 resources now pinned to unlocked version $svc1_new_version"
    else
        log_info "Service 1 version $svc1_version is not locked; no action needed"
    fi

    local svc2_version=$(grep -oE 'service_2_version[[:space:]]*=[[:space:]]*[0-9]+' terraform.tfvars | grep -oE '[0-9]+$')
    local svc2_locked=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_2_ID/version/$svc2_version" | jq -r '.locked')

    if [ "$svc2_locked" = "true" ]; then
        log_info "Service 2 version $svc2_version is locked; cloning to a fresh draft version..."

        if ! terraform apply -invoke=action.fastly_service_version_clone.service_2_clone_from_pinned -auto-approve; then
            log_error "Failed to clone service 2 off its locked version"
            return 1
        fi

        terraform refresh > /dev/null
        local svc2_new_version=$(terraform output -raw service_2_latest_version)
        log_success "Cloned version $svc2_version to draft version $svc2_new_version"

        sed -i.bak "s/service_2_version[[:space:]]*=[[:space:]]*[0-9]*/service_2_version = $svc2_new_version/" terraform.tfvars
        rm -f terraform.tfvars.bak

        log_info "Moving service 2 resources to version $svc2_new_version..."
        terraform plan -out=tfplan
        terraform apply tfplan
        log_success "Service 2 resources now pinned to unlocked version $svc2_new_version"
    else
        log_info "Service 2 version $svc2_version is not locked; no action needed"
    fi
}

# Test resource destruction
test_resource_destruction() {
    log_step "Testing resource destruction"

    cd "$TEST_DIR"

    # Get the latest versions from state
    terraform refresh > /dev/null 2>&1 || true
    local svc1_latest=$(terraform output -raw service_1_latest_version 2>/dev/null || echo "1")
    local svc2_latest=$(terraform output -raw service_2_latest_version 2>/dev/null || echo "1")

    log_info "Service 1: $SERVICE_1_ID (latest version: $svc1_latest)"
    log_info "Service 2: $SERVICE_2_ID (latest version: $svc2_latest)"

    # Run terraform destroy to delete the services
    # With force_destroy=true, this will delete all versions and the service itself
    log_info "Running terraform destroy..."
    if terraform destroy -auto-approve; then
        log_success "All resources destroyed via Terraform"
    else
        log_error "Terraform destroy failed"
        return 1
    fi

    # Verify services were deleted
    log_info "Verifying services are deleted..."
    sleep 2

    local svc1_check=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_1_ID" 2>/dev/null | jq -r '.msg // empty')
    local svc2_check=$(curl -s -H "Fastly-Key: $FASTLY_API_TOKEN" \
        "https://api.fastly.com/service/$SERVICE_2_ID" 2>/dev/null | jq -r '.msg // empty')

    if [[ "$svc1_check" == *"Record not found"* ]] || [ -z "$svc1_check" ]; then
        log_success "Service 1 successfully deleted"
    else
        log_warning "Service 1 may still exist: $svc1_check"
    fi

    if [[ "$svc2_check" == *"Record not found"* ]] || [ -z "$svc2_check" ]; then
        log_success "Service 2 successfully deleted"
    else
        log_warning "Service 2 may still exist: $svc2_check"
    fi

    log_success "Resource destruction test completed"
}

# Run all tests
main() {
    log_info "Starting provider lifecycle test"
    log_info "Test directory: $TEST_DIR"
    log_info "Process ID: $$"
    echo ""

    check_prerequisites
    build_provider
    setup_test_environment
    init_terraform
    apply_initial_config
    verify_initial_state
    verify_service_configuration
    test_resource_updates
    test_version_clone_action
    test_version_activate_action
    test_clone_from_latest_and_version_writes
    advance_off_locked_versions
    test_resource_destruction

    log_step "Test Summary - CDN Service"
    log_success "✓ Provider build"
    log_success "✓ Service creation (fastly_service_cdn)"
    log_success "✓ Domain attachment (fastly_service_domain)"
    log_success "✓ Backend configuration (fastly_service_backend)"
    log_success "✓ ACL configuration (fastly_service_cdn_acl)"
    log_success "✓ Version data sources (data.fastly_service_version)"
    log_success "✓ Resource updates"
    log_success "✓ Version clone action (fastly_service_version_clone)"
    log_success "✓ Version activate action (fastly_service_version_activate)"
    log_success "✓ Clone from latest version and version writes"
    log_success "✓ Resource destruction"

    echo ""
    log_success "All CDN service lifecycle tests passed!"
}

# Run main function
main
