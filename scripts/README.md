# Security Scanner Scripts

This directory contains automated security scanning tools for the `github.com/patdeg/common` repository.

## Quick Start

```bash
# Install security tools
./scripts/run-security-scans.sh --install

# Run all security scans
./scripts/run-security-scans.sh
```

## Scripts

### run-security-scans.sh

Comprehensive security scanner that runs multiple security analysis tools.

**Features:**
- Static Application Security Testing (SAST) with gosec
- Code quality checks with go vet and staticcheck
- Vulnerability detection with govulncheck
- Secret scanning for hardcoded credentials
- Dependency security audit
- Test coverage analysis
- Manual security pattern checks

**Usage:**

```bash
# Run all scans
./scripts/run-security-scans.sh

# Install required tools first
./scripts/run-security-scans.sh --install

# Install and run
./scripts/run-security-scans.sh --install --all
```

**Command-line Options:**

- `--install` - Install/update all security scanning tools
- `--all` - Run all available scans (including optional ones)

## Tools Used

### 1. gosec (Required)
**Static Application Security Testing (SAST)**

Scans Go code for common security issues:
- SQL injection vulnerabilities
- Command injection risks
- Hardcoded credentials
- Weak cryptography
- Insecure random number generation
- Path traversal vulnerabilities

**Installation:**
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

**Manual usage:**
```bash
gosec ./...
gosec -fmt=json -out=report.json ./...
```

### 2. govulncheck (Recommended)
**Vulnerability Database Scanner**

Checks dependencies against the Go vulnerability database.

**Installation:**
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

**Manual usage:**
```bash
govulncheck ./...
```

### 3. staticcheck (Optional)
**Advanced Static Analysis**

Performs advanced code quality and bug detection.

**Installation:**
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**Manual usage:**
```bash
staticcheck ./...
```

### 4. go vet (Built-in)
**Standard Go Analysis**

Built-in Go tool for detecting suspicious code constructs.

**Manual usage:**
```bash
go vet ./...
```

## Reports

All scan reports are saved to `security-reports/` directory with timestamps:

```
security-reports/
├── gosec-report-20251118_153045.json
├── gosec-summary-20251118_153045.txt
├── govet-report-20251118_153045.txt
├── staticcheck-report-20251118_153045.txt
├── govulncheck-report-20251118_153045.txt
├── secrets-check-20251118_153045.txt
├── dependencies-20251118_153045.txt
├── coverage-20251118_153045.txt
├── coverage-20251118_153045.html
└── manual-checks-20251118_153045.txt
```

### Report Types

1. **gosec-report-*.json** - Detailed JSON report of security issues
2. **gosec-summary-*.txt** - Human-readable summary
3. **govet-report-*.txt** - Go vet findings
4. **staticcheck-report-*.txt** - Staticcheck analysis
5. **govulncheck-report-*.txt** - Known vulnerabilities
6. **secrets-check-*.txt** - Hardcoded secrets scan
7. **dependencies-*.txt** - Full dependency list
8. **coverage-*.html** - Test coverage visualization
9. **manual-checks-*.txt** - Pattern-based security checks

## Interpreting Results

### gosec Severity Levels

- **HIGH** - Critical security issues requiring immediate attention
- **MEDIUM** - Moderate security concerns
- **LOW** - Minor security improvements

**Target:** 0 high severity issues

### Coverage Target

**Target:** >80% test coverage for security-critical code

## Web-Based Security Scans

After deploying your application, run these additional scans:

### 1. Mozilla Observatory
**URL:** https://observatory.mozilla.org/

Analyzes HTTP security headers and TLS configuration.

**Target Score:** B or higher

**Checks:**
- Content Security Policy
- Strict-Transport-Security
- X-Frame-Options
- X-Content-Type-Options
- And more...

### 2. Security Headers
**URL:** https://securityheaders.com/

Scans HTTP response headers for security best practices.

**Target Score:** B or higher

### 3. OWASP ZAP
**Dynamic Application Security Testing (DAST)**

Requires deployed application.

**Docker usage:**
```bash
# Baseline scan (passive)
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t http://your-app.com

# Full scan (active - may trigger security alerts)
docker run -t owasp/zap2docker-stable zap-full-scan.py \
  -t http://your-app.com
```

### 4. SSL Labs
**URL:** https://www.ssllabs.com/ssltest/

Tests SSL/TLS configuration for HTTPS sites.

**Target Grade:** A

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install security tools
        run: ./scripts/run-security-scans.sh --install

      - name: Run security scans
        run: ./scripts/run-security-scans.sh

      - name: Upload reports
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: security-reports
          path: security-reports/
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running security scans..."
./scripts/run-security-scans.sh

if [ $? -ne 0 ]; then
    echo "Security scan failed. Commit aborted."
    exit 1
fi
```

## Troubleshooting

### Tool Not Found

If you get "command not found" errors:

```bash
# Ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or install tools manually
./scripts/run-security-scans.sh --install
```

### jq Not Available

The script uses `jq` for JSON parsing. Install it:

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Alpine
apk add jq
```

### Scan Takes Too Long

For faster scans, exclude test files and vendor directories:

```bash
gosec -exclude-generated -exclude-dir=vendor ./...
```

## Best Practices

1. **Run scans before every commit**
   - Catches issues early
   - Prevents security debt

2. **Review all HIGH severity findings**
   - Address immediately
   - Document if false positive

3. **Update tools regularly**
   ```bash
   ./scripts/run-security-scans.sh --install
   ```

4. **Keep dependencies up-to-date**
   ```bash
   go get -u ./...
   go mod tidy
   ```

5. **Scan on every PR**
   - Automated via CI/CD
   - Blocks merging on failures

6. **Run full scans weekly**
   - Scheduled CI/CD job
   - Catches new vulnerabilities

## Security Issue Response

If security issues are found:

1. **High Severity:**
   - Fix immediately
   - Don't commit until resolved
   - Document the fix

2. **Medium Severity:**
   - Create GitHub issue
   - Fix within 1 week
   - Track in security backlog

3. **Low Severity:**
   - Create GitHub issue
   - Fix within 1 month
   - Consider in next sprint

4. **False Positives:**
   - Document why it's safe
   - Add to gosec exclusions if needed
   ```go
   // #nosec G101 -- This is not a credential
   const apiURL = "https://api.example.com"
   ```

## Additional Resources

- [OWASP Go Secure Coding Practices](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [gosec Documentation](https://github.com/securego/gosec)
- [Go Vulnerability Database](https://vuln.go.dev/)
- [NIST Secure Software Development Framework](https://csrc.nist.gov/Projects/ssdf)

## Maintenance

This script is maintained as part of the security remediation efforts for this repository.

**Last Updated:** 2025-11-18

For issues or improvements, please create a GitHub issue.
