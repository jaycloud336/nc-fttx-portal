## NC FTTX Portal - Automated CI Pipeline & Artifact Engineering

### Continuous Integration Repository
This repository manages the Continuous Integration (CI) for the NC FTTX Portal, a Go-based telecom infrastructure application. It is engineered to transform raw source code into Kubernetes-ready container images. By isolating the CI process here, every artifact is verified, tested, and scanned before it reaches the deployment phase.

### Continuous Deployment Repository
There is a second repository associated with this project. It can be found here:

`https://github.com/jaycloud336/nc-fttx-portal-gitops`

The `nc-fttx-portal-gitops` repository handles the Continuous Deployment process. Once the image is pushed to Docker Hub and the manifest is updated, ArgoCD detects the new image tag and automatically synchronizes the Kubernetes cluster state. All of the Kubernetes and ArgoCD manifest files are located in the Continuous Deployment repository.

### Major DevOps Components for artifact management:

**Artifact Hardening:** Multi-stage builds reduce image size, produce a minimal Go binary, and lower the overall attack surface.

**Security Scanning:** A separate security workflow runs dependency scanning (Trivy SCA), container image scanning (Trivy), and dynamic testing (OWASP ZAP) to detect vulnerabilities before and after the image is built. A dedicated CodeQL workflow adds static analysis (SAST) of the Go source, with findings posted to the repository Security tab.

**CI-to-CD Handshake:** Automated logic updates the companion GitOps (CD) repository once a new verified image is pushed.

### CI Workflow

**Developer (Push) ➡️ GitHub Actions (CI) ➡️ Multi-Stage Build ➡️ Push to Docker Hub ➡️ GitOps Manifest Update ➡️ ArgoCD Sync**

* **Optimized Build:** Multi-stage Docker build to package the application efficiently.
* **Registry Promotion:** Pushes the verified image to Docker Hub tagged with the branch name, commit SHA (`sha-<short>`), and `latest`.
* **Smoke Test:** Runs the built image and checks the `/health` endpoint before updating the manifest.
* **GitOps Trigger:** Updates the deployment manifest in the companion CD repository, which ArgoCD then synchronizes.

> **Note:** Security scanning runs as its own workflow (`security-scan.yml`) on pushes and pull requests to `main`, separate from the build/deploy pipeline. Static analysis runs in a third workflow (`codeql.yml`). See the guides below.

### Get Started - Clone or Fork the Project Repo

Follow the guides in this repository:

* **Verification Guide:** preview the working application (`Verifcation-Guide.md`)
* **CI Build & Push Pipeline:** walk through the build/deploy workflow (`cicd-automation-ci-artifact-build.md`)
* **Security Scan Pipeline:** walk through the security workflow (`cicd-automation-security-scan.md`)
* **Build-from-Scratch Implementation Guide:** create your own repo and directory structure from scratch (`cicd-implementation-guide.md`)