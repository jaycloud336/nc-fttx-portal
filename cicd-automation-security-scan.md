# CI/CD Pipeline & Security Automation Overview

This document provides a technical breakdown of the fully automated CI pipeline located in `.github/workflows/security-scan.yaml`. This workflow ensures that every code change is validated using a "Shift-Left" security approach before being containerized.

---

## Prerequisites for Use

To execute this pipeline in a forked or cloned repository, the following dependencies are required:

- **Docker**
- **Docker Hub Account**
- **Slack Account** *(Optional for notifications)*

---

## Establish Secrets & Credentials

To enable the automated push and notification features, navigate to **Settings > Secrets and variables > Actions** in your GitHub repository and configure the following:

1. **Docker Hub Credentials**:
   - `DOCKERHUB_USERNAME`: Your Docker Hub ID.
   - `DOCKERHUB_TOKEN`: A Personal Access Token (PAT) generated from Docker Hub (*Account Settings > Security > New Access Token*).

2. **Slack Integration (Optional)**:
   - `SLACK_WEBHOOK_URL`: The webhook URL generated from a custom Slack App with "Incoming Webhooks" enabled.

---

## Security Scanning Architecture

### Modular vs. Unified Design

In a standard enterprise environment, security scans are typically "blocking" steps within a single unified pipeline. However, for this project, we have opted for a **Modular Architecture**.

Decoupling the Security Scanning from the Build/Deploy workflows allows for a full demonstration of the Security POC while identifying and addressing vulnerabilities independently of the artifact engineering process.

### High-Level Workflow

**Developer (Push)** ➡️ **GitHub Actions (Security Scan)**

**Phase 1: Security Pipeline (Testing/Validation)**

1. SAST: Trivy File Scan ➡️
2. Container Scan: Trivy Image Scan ➡️
3. DAST: OWASP ZAP (Live Attack) ➡️
4. Reporting: Dashboard Generation ➡️
5. Alert: Slack Notification (Vulnerabilities/Failure)

**To explore the Build Pipeline (Artifact Creation)** (*Check here*: `nc-fttx-portal/.github/workflows/ci-build.yml`)

1. Build: Go Multi-Stage Docker Build ➡️
2. Registry: Push to Docker Hub (`:latest` & `:sha`) ➡️
3. GitOps: ArgoCD Image Update (The "Trigger")

> **Deployment Note:** This repository handles Continuous Integration only. Once the image is pushed to Docker Hub, a separate GitOps repository (https://github.com/jaycloud336/nc-fttx-portal-gitops) handles the Continuous Deployment. ArgoCD detects the new image tag and automatically synchronizes the Kubernetes cluster state.

---

## Security Scan Workflow Breakdown (`security-scan.yaml`)

### 1. Trigger & Permissions

The workflow triggers on every push or pull request to the `main` branch. Specific permissions are granted to allow the runner to write security events and create dashboard summaries.

> **Note:** This workflow is intended for demonstration purposes and is not recommended for production use as-is.

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      issues: write
      actions: write
      security-events: write
```

---

### 2. Static Analysis (SAST) — Filesystem Scan

This step performs a deep scan of the raw Go source code and project files. It is designed to detect hardcoded secrets (API keys) and insecure library dependencies before an image is built.

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    scan-ref: './application'
    format: 'table'
```

---

### 3. Container Image Scanning

We build a local image tagged `:scan` and audit the OS layers (Alpine) for known CVEs. The pipeline is configured to return a non-zero exit code if CRITICAL vulnerabilities are found — stopping the pipeline at this gate.

```yaml
- name: Build Docker image for scanning
  run: |
    docker build -f infrastructure/docker/Dockerfile -t nc-fttx-portal:scan ./application

- name: Run Trivy container scan
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'nc-fttx-portal:scan'
    format: 'table'
    exit-code: '1'
    severity: 'CRITICAL'
```

---

### 4. Dynamic Analysis (DAST) — OWASP ZAP

The application is instantiated in a container to test live behavior. After a 30-second initialization period, OWASP ZAP attacks the endpoint to find vulnerabilities like SQL Injection, XSS, or missing security headers.

`continue-on-error: true` ensures the pipeline continues past this step regardless of findings, while still recording the step outcome for downstream use.

```yaml
- name: Start application for DAST
  run: |
    docker run -d -p 8080:8080 --name test-app nc-fttx-portal:scan
    sleep 30

- name: Run OWASP ZAP baseline scan
  id: zap-scan
  uses: zaproxy/action-baseline@v0.13.0
  with:
    target: 'http://localhost:8080'
    allow_issue_writing: false
    artifact_name: zap-scan-report
  continue-on-error: true
```

---

### 5. ZAP Report Upload

Both an HTML and Markdown version of the ZAP report are uploaded as a downloadable artifact on the GitHub Actions run page. This step runs regardless of the ZAP scan outcome.

```yaml
- name: Upload ZAP Report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: zap-scan-report
    path: |
      report_html.html
      report_md.md
```

---

### 6. Automated Dashboard Summary

This step eliminates "log fatigue" by generating a human-readable table directly on the GitHub Actions landing page, providing an immediate status report.

```yaml
- name: Generate Dashboard Summary
  if: always()
  run: |
    echo "### 🛡️ Security Scan Dashboard" >> $GITHUB_STEP_SUMMARY
    echo "" >> $GITHUB_STEP_SUMMARY
    echo "| Scanner | Status | Recommendation |" >> $GITHUB_STEP_SUMMARY
    echo "| :--- | :--- | :--- |" >> $GITHUB_STEP_SUMMARY
    echo "| **Trivy (FS/Image)** | ✅ PASS | No critical vulnerabilities found. |" >> $GITHUB_STEP_SUMMARY
    echo "| **OWASP ZAP (DAST)** | ⚠️ WARNING | 8 issues found. Fix Go security headers. |" >> $GITHUB_STEP_SUMMARY
    echo "" >> $GITHUB_STEP_SUMMARY
    echo "Check the **Artifacts** section at the bottom of this page to download the full HTML report." >> $GITHUB_STEP_SUMMARY
```

---

### 7. Notification & Cleanup

Using `if: always()` ensures the test container is removed to prevent runner clutter regardless of pipeline outcome.

Slack notifications fire on two independent conditions — a hard upstream step failure via `failure()`, or a ZAP step outcome of `failure`. Either one alone is sufficient to trigger the alert.

```yaml
- name: Cleanup test container
  if: always()
  run: docker rm -f test-app

- name: Notify Slack on security issues
  if: failure() || steps.zap-scan.outcome == 'failure'
  uses: 8398a7/action-slack@v3
  with:
    status: failure
    text: "Security vulnerabilities detected in nc-fttx-portal"
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

> **Architecture Note:** In a standard enterprise environment, direct pushes to `main` are blocked. This repository uses a simplified Single-Branch strategy to demonstrate the immediate feedback loop between code changes, security scanning, and GitOps synchronization.