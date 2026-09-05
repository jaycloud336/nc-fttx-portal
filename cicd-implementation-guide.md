# CI/CD Implementation Guide - NC FTTX Portal

## Creation Of GitHub Repository

- Repository name: `nc-fttx-portal`
- Description: `DevOps CI/CD Telecom Infrastructure Permitting Portal`
- Click **"Create Repository"**

### Step 2: Initialize Local Git Repository in proper directory

Navigate to your project directory and initialize git with an initial commit:

```bash
cd ~/your/path/nc-fttx-portal
git init
git add .
git commit -m "Initial commit: NC FTTX Portal with containerization"
```

Add the GitHub remote and push to main:

```bash
git remote add origin https://github.com/YOUR_USERNAME/nc-fttx-portal.git
git branch -M main
git push -u origin main
```

### Step 3: Verify Repository Structure

Workflows live in `.github/workflows/`.

```
nc-fttx-portal/
├── .github/
│   └── workflows/
│       ├── ci-build.yml           # Build, test, push
│       └── security-scan.yml      # Security scanning workflow
│       └── codeql.yml             # CodeQL SAST (static code analysis)
├── application/
│   ├── web/
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── infrastructure/
│   └── docker/
│       └── Dockerfile
├── docs/
├── .gitignore
├── cicd-implementation-guide.md
└── README.md
```

---

## Phase 2: GitHub Actions CI Pipeline Configuration

**Establish Docker Hub Credentials**

Docker Hub Website, **Create Access Token:**
- Go to Account Settings > Security
- Click "New Access Token"
- Name: `gh-actions-nc-fttx`
- Permissions: **Read & Write**
- Copy the token (save it securely). Example: `dckr_pat_<YOUR_TOKEN_HERE>`

GitHub Website, **Configure Repository Secrets**
1. Go to your GitHub repository
2. **Settings > Secrets and variables > Actions**
3. Add the following repository secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub access token
   - `PERSONAL_ACCESS_TOKEN`: A GitHub PAT with `repo` (or `public_repo`) scope, allowing the runner to commit to the GitOps repository

> **Token expiry:** Both the Docker Hub token and the GitHub PAT expire. If the pipeline fails at "Login to Docker Hub" or "Checkout GitOps repository," regenerate the relevant token and update its secret.

---

## Phase 3: Create CI Pipeline Workflow

Create the file at `.github/workflows/ci-build.yml`.

The workflow triggers on pushes to `main` under the application and Dockerfile paths. The `pull_request` block is included but commented out as an optional safety gate for a multi-branch team environment.

```yaml
name: CI Build and Push

on:
  push:
    branches: [main]
    paths:
      - 'application/**'
      - 'infrastructure/docker/Dockerfile'
  # Optional code for multi-branch environments
  # pull_request:
  #   branches: [main]
  #   paths:
  #     - 'application/**'

env:
  REGISTRY: docker.io
  IMAGE_NAME: nc-fttx-portal
  GITOPS_REPO_PATH: 'nc-fttx-portal-gitops'
  DEPLOYMENT_MANIFEST_PATH: 'manifests/deployment.yaml'

jobs:
  build-and-push:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Login to Docker Hub
      if: github.event_name != 'pull_request'
      uses: docker/login-action@v3
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix=sha-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: ./application
        file: ./infrastructure/docker/Dockerfile
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Test container
      if: github.event_name != 'pull_request'
      run: |
        docker run -d -p 8080:8080 --name test-container ${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}:latest
        sleep 10
        curl --retry 5 --retry-delay 2 --fail http://localhost:8080/health || (docker logs test-container && exit 1)
        docker stop test-container
        docker rm test-container

  update-manifest:
    runs-on: ubuntu-latest
    needs: build-and-push
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
    - name: Checkout GitOps repository
      uses: actions/checkout@v4
      with:
        repository: jaycloud336/nc-fttx-portal-gitops
        path: ${{ env.GITOPS_REPO_PATH }}
        token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}

    - name: Update image tag in deployment manifest
      run: |
        cd ${{ env.GITOPS_REPO_PATH }}
        SHORT_SHA=$(echo "${{ github.sha }}" | cut -c1-7)
        NEW_TAG="${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}:sha-${SHORT_SHA}"
        sed -i 's|image: .*/nc-fttx-portal:.*|image: '"$NEW_TAG"'|g' ${{ env.DEPLOYMENT_MANIFEST_PATH }}

    - name: Commit and push changes
      uses: EndBug/add-and-commit@v9
      with:
        message: 'chore: update application image to sha-${{ github.sha }} [skip ci]'
        cwd: ${{ env.GITOPS_REPO_PATH }}
        add: '${{ env.DEPLOYMENT_MANIFEST_PATH }}'
        default_author: github_actions
```

---

## Phase 4: Security Scanning Integration

Create the file at `.github/workflows/security-scan.yml`.

This workflow runs on every push and pull request to `main` (shift-left), layering dependency scanning (Trivy SCA), container image scanning (Trivy image), and dynamic testing (OWASP ZAP). The Trivy action is pinned to a release rather than a moving branch, the app is polled for readiness rather than waited on with a fixed sleep, and the dashboard reads its values live from the scan results.

```yaml
name: Security Scanning

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

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@v0.35.0
      with:
        scan-type: 'fs'
        scan-ref: './application'
        format: 'table'

    - name: Build Docker image for scanning
      run: |
        docker build -f infrastructure/docker/Dockerfile -t nc-fttx-portal:scan ./application

    - name: Run Trivy container scan
      id: trivy-image
      uses: aquasecurity/trivy-action@v0.35.0
      with:
        image-ref: 'nc-fttx-portal:scan'
        format: 'table'
        exit-code: '1'
        severity: 'CRITICAL'

    - name: Start application for DAST
      run: |
        docker run -d -p 8080:8080 --name test-app nc-fttx-portal:scan
        for i in $(seq 1 30); do
          curl -sf http://localhost:8080/health && break
          sleep 2
        done

    - name: Run OWASP ZAP baseline scan
      id: zap-scan
      uses: zaproxy/action-baseline@v0.14.0
      with:
        target: 'http://localhost:8080'
        allow_issue_writing: false
        artifact_name: zap-scan-report
      continue-on-error: true

    - name: Cleanup test container
      if: always()
      run: docker rm -f test-app

    - name: Generate Dashboard Summary
      if: always()
      run: |
        if [ -f report_json.json ]; then
          HIGH=$(jq '[.site[].alerts[] | select(.riskcode=="3")] | length' report_json.json)
          MED=$(jq '[.site[].alerts[] | select(.riskcode=="2")] | length' report_json.json)
          LOW=$(jq '[.site[].alerts[] | select(.riskcode=="1")] | length' report_json.json)
          ZAP_SUMMARY="High: ${HIGH}, Medium: ${MED}, Low: ${LOW}"
        else
          ZAP_SUMMARY="Report not found"
        fi

        if [ "${{ steps.trivy-image.outcome }}" = "success" ]; then
          TRIVY_STATUS="PASS"
          TRIVY_NOTE="No critical vulnerabilities found."
        else
          TRIVY_STATUS="FAIL"
          TRIVY_NOTE="Critical vulnerabilities detected, see logs."
        fi

        echo "### Security Scan Dashboard" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "| Scanner | Status | Findings |" >> $GITHUB_STEP_SUMMARY
        echo "| :--- | :--- | :--- |" >> $GITHUB_STEP_SUMMARY
        echo "| Trivy (Image) | ${TRIVY_STATUS} | ${TRIVY_NOTE} |" >> $GITHUB_STEP_SUMMARY
        echo "| OWASP ZAP (DAST) | WARNING | ${ZAP_SUMMARY} |" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "Check the Artifacts section at the bottom of this page to download the full HTML report." >> $GITHUB_STEP_SUMMARY

    - name: Notify Slack on security issues
      if: failure() || steps.zap-scan.outcome == 'failure'
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        text: "Security vulnerabilities detected in nc-fttx-portal"
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

> Example webhook format: `https://hooks.slack.com/services/<YOUR_HOOK_HERE>`

---

## Phase 5: Testing and Deployment

### Step 8: Commit and Push the Workflows

```bash
git add .github/workflows/
git commit -m "Add GitHub Actions CI/CD pipeline with security scanning"
git push origin main
```

A push to `main` that touches `application/**` or the Dockerfile triggers the CI build. Any push to `main` triggers the security scan.

### Step 9: Monitor Pipeline Execution

1. Go to the GitHub repository
2. Click the **Actions** tab
3. Watch both workflows run: CI Build and Push, and Security Scanning
4. Check Docker Hub for the pushed image

---

## Phase 6: Verification

**Expected Results:**
- GitHub Actions shows green checkmarks for both workflows
- Docker Hub contains the image tagged with `main`, `latest`, and `sha-<short>`
- The `update-manifest` job commits the new SHA tag to the GitOps repo
- Trivy reports no critical vulnerabilities; ZAP report uploads as an artifact
- Container health check passes

### Artifacts
- Container images: `docker.io/jaycloud336/nc-fttx-portal`
- Tagged with git commit SHA, branch name, and `latest`

---

## Important Note: Production Architecture vs. Repo Implementation

In a standard enterprise environment, this pipeline would be distributed across multiple environments (Dev, Staging, Prod), and security scans would be blocking gates in a single unified pipeline. For this repo, a simplified single-branch strategy is used, with security decoupled into its own workflow, to demonstrate the immediate feedback loop between code changes and GitOps synchronization.

> **Continue to the Deployment Repo:** https://github.com/jaycloud336/nc-fttx-portal-gitops