## CI/CD Pipeline (GitHub Actions)

The project includes a fully automated CI pipeline located in `.github/workflows/ci.yaml`. This workflow ensures that every code change is validated and containerized before being deployed to the cluster.

### Key Pipeline Stages:

* **Security Scan**: Scans the Docker image for known vulnerabilities (CVEs) using Trivy.
* **Build & Push**: Compiles the Go application and pushes a tagged image to Docker Hub using the Git SHA for version immutability.
* **Slack Notification(Optional)**: Sends real-time build status updates to a designated Slack channel.


---

### To use this pipeline in your own forked or cloned Github repo you must have:

-Docker
-Dockerhub Account
-Slack Account(Optional)

## Establish Secrets/Credentials**

Docker Hub Website 
**Create Access Token:**
   - Go to Account Settings > Security
   - Click "New Access Token"
   - Name: `github-actions-nc-fttx`
   - Permissions: Read, Write, Delete
   - Copy the token (save it securely) *Example:*dckr_pat_<YOUR_TOKEN_HERE>*


**Configure Secrets**: In your GitHub Repository, navigate to **Settings > Secrets and variables > Actions** and add the following:
    * `DOCKERHUB_USERNAME`: Your Docker Hub ID.
    * `DOCKERHUB_TOKEN`: A Personal Access Token (PAT) generated from your Docker Hub account.

**Slack (Optional - Create Webhook):**
   - Create a Slack App in your workspace.
   - Enable **"Incoming Webhooks"** and create a new Webhook to a specific channel.
   - Copy the Webhook URL (starts with `https://hooks.slack.com/...`).



**Modify the Image Name**: Update the `tags:` section in the `.github/workflows/ci.yaml` file to reflect your own Docker Hub username instead of `jaycloud336`.


## Security Scanning

### For this repo the Security Scanning and Build/Deploy workflows are separated into two distinct YAML files.

In a high-maturity production environment, security scans are typically "blocking" steps within a single unified pipeline. However, for this project repo we have opted for a Modular Architecture.

Decoupling these processes allows for a full demonstration of the POC while still recognizing and addresing any vulnurabilites if they exist.


### The High-Level CI Workflow:

Developer (Push) ➡️ GitHub Actions (CI)

*Enter Security Pipeline for image validation*

SAST: Trivy File Scan ➡️

Container Scan: Trivy Image Scan ➡️

DAST: OWASP ZAP (Live Attack) ➡️

Alert: Slack Notification

*Enter Build Pipeline for artfact creation*

Build: Go Multi-Stage Docker Build ➡️

Registry: Push to Docker Hub (:latest & :sha) ➡️

GitOps: ArgoCD Image Update (The "Trigger")

***Note:** Note on Deployment: This repository (Repo 1) handles the Continuous Integration. Once the image is pushed to Docker Hub, the Continuous Deployment (Repo 2) is triggered. ArgoCD will detect the new image tag and automatically synchronize the Kubernetes cluster state.**

## Security Scan YAML

2. Trigger (Initiation Event)

```YAML
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read
  issues: write
  security-events: write
Triggers: Automatically runs on every push or PR to main.
```

**Permissions: Specifically grants the runner the ability to write to the GitHub Security tab and create issues if vulnerabilities are found.**


***Important Note:***

***Production Architecture vs. Repo Implementation***

*In a standard enterprise environment, this pipeline would be distributed across multiple environments (Dev, Staging, Prod). Typically developers will work on `feature/*` branches. Direct pushes to `main` are blocked. Code enters `main` only after a successful Peer Review and CI pass on a Pull Request. For the purpose of this Repo, a simplified Single-Branch (Main) strategy is used to demonstrate the immediate feedback loop between code changes, security scanning, and GitOps synchronization.*

2. Static Analysis (Filesystem Scan)
```YAML
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    scan-ref: './application'
```

**Performs a deep scan of the raw Go source code and project files. Detects hardcoded secrets (like API keys) and insecure library dependencies in the 'fs' file system before the image is even built.**

3. Container Image Scanning
```YAML
- name: Build Docker image for scanning
  run: |
    docker build -f infrastructure/docker/Dockerfile -t nc-fttx-portal:scan ./application

- name: Run Trivy container scan
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'nc-fttx-portal:scan'
```

**Builds a local image tagged :scan and runs Trivy against it. Scans the OS layers (like Alpine or Debian) for known vulnerabilities. This catches issues that exist in the base image and its base layers.**

4. Dynamic Analysis (DAST)
```YAML
- name: Start application for DAST
  run: |
    docker run -d -p 8080:8080 --name test-app nc-fttx-portal:scan
    sleep 30

- name: Run OWASP ZAP baseline scan
  uses: zaproxy/action-baseline@v0.13.0
  with:
    target: 'http://localhost:8080'
```

**Spins up the app, waits 30 seconds for it to initialize, and then hits it with OWASP ZAP. Purpose: Tests the running application for "live" vulnerabilities like SQL injection or cross-site scripting (XSS)**

5. Notification & Cleanup (Optional)

```YAML
- name: Cleanup test container
  if: always()
  run: docker rm -f test-app
  # (Optional) If you aren't using Slack, you should omit or comment out the folleing notification step entirely.
- name: Notify Slack on security issues
  if: failure()
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
    SLACK_MESSAGE: "Security scan failed in the CI pipeline!"
```
***SLACK: Only triggers if a previous step fails, ensuring you only get "noise" when there is a real security problem to fix. Using 'if: always()' ensures the container is deleted even if the scan fails, preventing runner clutter.***