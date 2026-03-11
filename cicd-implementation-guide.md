# CI/CD Implementation Guide - NC FTTX Portal

1. ## Creation Of GitHub Repository
   **GitHub Repository Established as "Single Source Of Truth**

   - Repository name: `nc-fttx-portal`
   - Description: `DevOps CI/CD - Telecom Infrastructure Permitting Portal`

4. **Click "Create Repository"**

### Step 2: Initialize Local Git Repository in proper directory

### Navigate to your project directory and initialize git folder initial commit
```bash
cd ~/desired directory path/nc-fttx-portal
git init
git add .
git commit -m "Initial commit: NC FTTX Portal with containerization"
```
### Add GitHub remote repo and establish proper branch for push
```bash
git remote add origin https://github.com/YOUR_USERNAME/nc-fttx-portal.git`
git branch -M main
git push -u origin main
```
### Step 3: Verify Repository Structure

**All necessary source code files should now be in GitHub repo:**

```
nc-fttx-portal/
├── application/              # Go application source code files
├── ci-pipeline/             # CI/CD configuration (can be constructed locally or in github GUI)
├── README.md               # Project documentation
└── .gitignore             # Git ignore patterns
```

```
nc-fttx-portal/
├── .github/
│   └── workflows/
│       ├── ci-build.yml           # Build, test, and push workflow
│       ├── security-scan.yml      # Security scanning workflow  
│       └── cd-deploy.yml          # CD trigger workflow (future)
├── application/
│   ├── web/
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── docs/
├── infrastructure/
│   └── docker/
│       └── Dockerfile
├── .gitignore
├── cicd-implementation-guide.md
└── README.md
```


---

## Phase 2: GitHub Actions CI Pipeline Configuration

**Establish Docker Hub Credentials**

Docker Hub Website 
**Create Access Token:**
   - Go to Account Settings > Security
   - Click "New Access Token"
   - Name: `github-actions-nc-fttx`
   - Permissions: Read, Write, Delete
   - Copy the token (save it securely) Example: *dckr_pat_<YOUR_TOKEN_HERE>*

Github Website
**Configure GitHub Secret**
1. **Go to your GitHub repository**
2. **Settings > Secrets and variables > Actions**
3. **Add Repository Secrets:**
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub access token

### Phase 3: Create CI Pipeline Workflow

Create file: `ci-pipeline/github-actions/ci-build.yml`

```yaml
name: CI Build and Push

on:
  push:
    branches: [main, develop]
    paths: ['application/**']
  pull_request:
    branches: [main]
    paths: ['application/**']

env:
  REGISTRY: docker.io
  IMAGE_NAME: nc-fttx-portal

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
        curl -f http://localhost:8080/health || exit 1
        docker stop test-container
        docker rm test-container
```

---

## Phase 3: Security Scanning Integration


Create file: `ci-pipeline/github-actions/security-scan.yml`

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
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: './application'
        format: 'table'

    - name: Build Docker image for scanning
      run: |
        docker build -f infrastructure/docker/Dockerfile -t nc-fttx-portal:scan ./application

    - name: Run Trivy container scan
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: 'nc-fttx-portal:scan'
        format: 'table'

    - name: Notify Slack on security issues
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        text: "Security vulnerabilities detected in nc-fttx-portal"
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```
### ***Example Webhook: https://hooks.slack.com/services/<YOUR_HOOK_HERE>****

---

## Phase 4: Testing and Deployment

### Step 8: Test the Pipeline Locally

Before pushing, test your workflow files:

```bash
# Validate workflow syntax (if you have act installed)
act -l

# Or push to a test branch first
git checkout -b test-pipeline
git add ci-pipeline/github-actions/
git commit -m "Add CI/CD pipeline workflows"
git push origin test-pipeline
```

### Step 9: Push Pipeline to Main Branch

```bash
# Switch back to main
git checkout main

# Add the workflow files
git add ci-pipeline/github-actions/

# Commit the pipeline
git commit -m "Add GitHub Actions CI/CD pipeline with security scanning"

# Push to trigger the pipeline
git push origin main
```

### Step 10: Monitor Pipeline Execution

1. **Go to GitHub repository**
2. **Click "Actions" tab**
3. **Watch the pipeline run:**
   - Build and Push workflow
   - Security scanning workflow
4. **Check Docker Hub** for pushed images

---

## Phase 5: Verification and Documentation

### Step 11: Verify Pipeline Success

**Expected Results:**
- GitHub Actions shows green checkmarks
- Docker Hub contains your image
- Security scans complete without critical issues
- Container health check passes

### Step 12: Update Documentation

Add to your README.md:

```markdown
## CI/CD Pipeline

This project includes automated CI/CD pipeline with:
- **Automated building** on code changes
- **Security scanning** with Trivy
- **Container registry** integration with Docker Hub
- **Health check validation** for deployments

### Pipeline Triggers
- Push to main/develop branches
- Pull requests to main branch
- Only triggers on application code changes

### Artifacts
- Container images available at: `docker.io/jaycloud336/nc-fttx-portal`
- Tagged with git commit SHA and branch names
```

---

