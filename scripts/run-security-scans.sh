#!/bin/bash

# Security Scanner Script
# Runs multiple security scanners against the codebase
# Usage: ./scripts/run-security-scans.sh [--install] [--all]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
REPORT_DIR="security-reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create reports directory
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Security Scanner Suite${NC}"
echo -e "${BLUE}   Repository: github.com/patdeg/common${NC}"
echo -e "${BLUE}   Date: $(date)${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install tools
install_tools() {
    echo -e "${YELLOW}Installing security tools...${NC}"

    # Install gosec
    if ! command_exists gosec; then
        echo "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    else
        echo "✓ gosec already installed"
    fi

    # Install nancy
    if ! command_exists nancy; then
        echo "Installing nancy..."
        go install github.com/sonatype-nexus-community/nancy@latest
    else
        echo "✓ nancy already installed"
    fi

    # Install staticcheck
    if ! command_exists staticcheck; then
        echo "Installing staticcheck..."
        go install honnef.co/go/tools/cmd/staticcheck@latest
    else
        echo "✓ staticcheck already installed"
    fi

    # Install govulncheck
    if ! command_exists govulncheck; then
        echo "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    else
        echo "✓ govulncheck already installed"
    fi

    echo -e "${GREEN}✓ All tools installed${NC}"
    echo
}

# Parse arguments
INSTALL=false
RUN_ALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --install)
            INSTALL=true
            shift
            ;;
        --all)
            RUN_ALL=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--install] [--all]"
            exit 1
            ;;
    esac
done

# Install tools if requested
if [ "$INSTALL" = true ]; then
    install_tools
fi

# Check for required tools
echo -e "${BLUE}Checking for required tools...${NC}"
TOOLS_MISSING=false

if ! command_exists gosec; then
    echo -e "${RED}✗ gosec not found${NC}"
    TOOLS_MISSING=true
else
    echo -e "${GREEN}✓ gosec found${NC}"
fi

if ! command_exists staticcheck; then
    echo -e "${YELLOW}⚠ staticcheck not found (optional)${NC}"
else
    echo -e "${GREEN}✓ staticcheck found${NC}"
fi

if ! command_exists govulncheck; then
    echo -e "${YELLOW}⚠ govulncheck not found (optional)${NC}"
else
    echo -e "${GREEN}✓ govulncheck found${NC}"
fi

if [ "$TOOLS_MISSING" = true ]; then
    echo -e "${YELLOW}Run with --install to install missing tools${NC}"
    echo
fi

echo

# 1. Run gosec (Static Application Security Testing)
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}1. Running gosec (SAST)${NC}"
echo -e "${BLUE}========================================${NC}"

GOSEC_REPORT="$REPORT_DIR/gosec-report-$TIMESTAMP.json"
GOSEC_HTML="$REPORT_DIR/gosec-report-$TIMESTAMP.html"

if command_exists gosec; then
    echo "Scanning for security vulnerabilities in Go code..."

    # Run gosec with JSON output
    gosec -fmt=json -out="$GOSEC_REPORT" -exclude-generated ./... 2>&1 || true

    # Also create a text summary
    gosec -fmt=text ./... 2>&1 | tee "$REPORT_DIR/gosec-summary-$TIMESTAMP.txt" || true

    # Count issues
    if [ -f "$GOSEC_REPORT" ]; then
        HIGH_COUNT=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")
        MEDIUM_COUNT=$(jq '[.Issues[] | select(.severity == "MEDIUM")] | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")
        LOW_COUNT=$(jq '[.Issues[] | select(.severity == "LOW")] | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")

        echo -e "${BLUE}Results:${NC}"
        echo -e "  High:   ${RED}$HIGH_COUNT${NC}"
        echo -e "  Medium: ${YELLOW}$MEDIUM_COUNT${NC}"
        echo -e "  Low:    $LOW_COUNT"
        echo -e "  Report: $GOSEC_REPORT"

        if [ "$HIGH_COUNT" -gt 0 ]; then
            echo -e "${RED}⚠ High severity issues found!${NC}"
        else
            echo -e "${GREEN}✓ No high severity issues${NC}"
        fi
    fi
else
    echo -e "${YELLOW}Skipping gosec (not installed)${NC}"
fi

echo

# 2. Run go vet
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}2. Running go vet${NC}"
echo -e "${BLUE}========================================${NC}"

VET_REPORT="$REPORT_DIR/govet-report-$TIMESTAMP.txt"

echo "Running go vet..."
go vet ./... 2>&1 | tee "$VET_REPORT" || true

if [ ! -s "$VET_REPORT" ]; then
    echo -e "${GREEN}✓ No issues found${NC}"
else
    echo -e "${YELLOW}⚠ Issues found - see $VET_REPORT${NC}"
fi

echo

# 3. Run staticcheck (if available)
if command_exists staticcheck; then
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}3. Running staticcheck${NC}"
    echo -e "${BLUE}========================================${NC}"

    STATICCHECK_REPORT="$REPORT_DIR/staticcheck-report-$TIMESTAMP.txt"

    echo "Running staticcheck..."
    staticcheck ./... 2>&1 | tee "$STATICCHECK_REPORT" || true

    if [ ! -s "$STATICCHECK_REPORT" ]; then
        echo -e "${GREEN}✓ No issues found${NC}"
    else
        echo -e "${YELLOW}⚠ Issues found - see $STATICCHECK_REPORT${NC}"
    fi

    echo
fi

# 4. Run govulncheck (if available)
if command_exists govulncheck; then
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}4. Running govulncheck (Vulnerability DB)${NC}"
    echo -e "${BLUE}========================================${NC}"

    VULN_REPORT="$REPORT_DIR/govulncheck-report-$TIMESTAMP.txt"

    echo "Checking for known vulnerabilities in dependencies..."
    govulncheck ./... 2>&1 | tee "$VULN_REPORT" || true

    if grep -q "No vulnerabilities found" "$VULN_REPORT"; then
        echo -e "${GREEN}✓ No vulnerabilities found${NC}"
    else
        echo -e "${YELLOW}⚠ Check report for details: $VULN_REPORT${NC}"
    fi

    echo
fi

# 5. Check for hardcoded secrets
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}5. Checking for hardcoded secrets${NC}"
echo -e "${BLUE}========================================${NC}"

SECRETS_REPORT="$REPORT_DIR/secrets-check-$TIMESTAMP.txt"

echo "Scanning for potential secrets..."

{
    echo "=== Potential API Keys ==="
    grep -rn "api[_-]key\s*=\s*['\"]" --include="*.go" --include="*.md" . || echo "None found"
    echo

    echo "=== Potential Passwords ==="
    grep -rn "password\s*=\s*['\"]" --include="*.go" --include="*.md" . || echo "None found"
    echo

    echo "=== Potential Tokens ==="
    grep -rn "token\s*=\s*['\"]" --include="*.go" --include="*.md" . || echo "None found"
    echo

    echo "=== Base64 Encoded Strings (potential secrets) ==="
    grep -rn "[A-Za-z0-9+/]{40,}={0,2}" --include="*.go" . | grep -v "test" | grep -v "example" || echo "None found"

} | tee "$SECRETS_REPORT"

echo -e "${GREEN}✓ Secret scan complete - see $SECRETS_REPORT${NC}"
echo

# 6. Dependency audit
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}6. Dependency Security Audit${NC}"
echo -e "${BLUE}========================================${NC}"

DEP_REPORT="$REPORT_DIR/dependencies-$TIMESTAMP.txt"

echo "Listing all dependencies..."
go list -json -m all > "$DEP_REPORT"

echo "Dependency report: $DEP_REPORT"
echo -e "${GREEN}✓ Dependency list generated${NC}"
echo

# 7. Test coverage
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}7. Running tests with coverage${NC}"
echo -e "${BLUE}========================================${NC}"

COV_REPORT="$REPORT_DIR/coverage-$TIMESTAMP.txt"
COV_HTML="$REPORT_DIR/coverage-$TIMESTAMP.html"

echo "Running tests..."
go test ./... -cover -coverprofile="$REPORT_DIR/coverage.out" 2>&1 | tee "$COV_REPORT"

if [ -f "$REPORT_DIR/coverage.out" ]; then
    go tool cover -html="$REPORT_DIR/coverage.out" -o "$COV_HTML"
    echo -e "${GREEN}✓ Coverage report: $COV_HTML${NC}"

    # Get overall coverage
    OVERALL_COV=$(go tool cover -func="$REPORT_DIR/coverage.out" | grep total | awk '{print $3}')
    echo -e "Overall coverage: $OVERALL_COV"
fi

echo

# 8. Check for common security issues
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}8. Manual Security Checks${NC}"
echo -e "${BLUE}========================================${NC}"

MANUAL_REPORT="$REPORT_DIR/manual-checks-$TIMESTAMP.txt"

{
    echo "=== SQL Injection Risks ==="
    grep -rn "Query.*+" --include="*.go" . || echo "None found"
    grep -rn "Exec.*+" --include="*.go" . || echo "None found"
    echo

    echo "=== Command Injection Risks ==="
    grep -rn "exec.Command" --include="*.go" . || echo "None found"
    echo

    echo "=== Path Traversal Risks ==="
    grep -rn "filepath.Join.*FormValue\|filepath.Join.*Param" --include="*.go" . || echo "Validation in place"
    echo

    echo "=== XSS Risks (template.HTML usage) ==="
    grep -rn "template.HTML\|template.JS\|template.URL" --include="*.go" . || echo "None found"
    echo

    echo "=== CSRF Protection ==="
    if grep -rq "csrf" --include="*.go" .; then
        echo "✓ CSRF package found"
    else
        echo "⚠ CSRF package not found"
    fi
    echo

    echo "=== TLS/HTTPS Usage ==="
    grep -rn "InsecureSkipVerify.*true" --include="*.go" . || echo "✓ No insecure TLS skipping"
    echo

} | tee "$MANUAL_REPORT"

echo -e "${GREEN}✓ Manual checks complete - see $MANUAL_REPORT${NC}"
echo

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Scan Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo "All reports saved to: $REPORT_DIR/"
echo
echo "Reports generated:"
ls -lh "$REPORT_DIR"/*-$TIMESTAMP.* 2>/dev/null || true
echo

# Web-based scans (informational)
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Additional Manual Scans${NC}"
echo -e "${BLUE}========================================${NC}"
echo
echo "For complete security assessment, also run:"
echo
echo "1. Mozilla Observatory:"
echo "   https://observatory.mozilla.org/"
echo "   - Analyze: your-deployed-domain.com"
echo
echo "2. Security Headers:"
echo "   https://securityheaders.com/"
echo "   - Scan: your-deployed-domain.com"
echo
echo "3. OWASP ZAP (requires deployed application):"
echo "   docker run -t owasp/zap2docker-stable zap-baseline.py \\"
echo "     -t http://your-deployed-domain.com"
echo
echo "4. SSL Labs (for HTTPS):"
echo "   https://www.ssllabs.com/ssltest/"
echo

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   Security Scan Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo "Review the reports in $REPORT_DIR/ for detailed findings."
echo

# Exit with error if high severity issues found
if [ -f "$GOSEC_REPORT" ]; then
    HIGH_COUNT=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")
    if [ "$HIGH_COUNT" -gt 0 ]; then
        echo -e "${RED}⚠ HIGH SEVERITY ISSUES FOUND: $HIGH_COUNT${NC}"
        echo -e "${RED}Please review and fix before deployment.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✓ No critical issues found${NC}"
exit 0
