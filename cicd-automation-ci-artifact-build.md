## CI Build & Push Pipeline (GitHub Actions)

The project includes a fully automated CI pipeline located in `.github/workflows/ci.yaml`. This workflow ensures that every code change is validated and containerized before being deployed to the cluster.

### Key Pipeline Stages:

* **Metadata Extraction**: Dynamically generates image tags using the Git SHA (Short) and branch names to ensure every build is unique and traceable.
* **Build & Push**: Compiles the Go application and pushes a tagged image to Docker Hub using the Git SHA for version immutability.
* **Smoke Test (Health Check)**: Runs the newly built image in the runner to verify the app starts and responds correctly before updating the manifest.
* **GitOps Update**: Automatically updates the image tag in the separate GitOps repository to trigger an ArgoCD synchronization.

---

### To use this pipeline in your own forked or cloned repo you must have:

* Docker Hub Account
* GitHub Personal Access Token (PAT) with repo permissions
* Access to the corresponding GitOps Repository


## ***Shared Credentials***
***Note: If you have already configured your Docker Hub tokens and GitHub Secrets for the "Security Scanning" pipeline, you can skip the "Establish Secrets/Credentials" section below. These credentials are encrypted at the repository level and are automatically available to both pipelines.***

---

## ***Establish Secrets/Credentials*** **(Only if needed)**

**Docker Hub Website**
**Create Access Token:**
   - Go to Account Settings > Security
   - Click "New Access Token"
   - Name: `github-actions-nc-fttx`
   - Permissions: Read, Write, Delete
   - Copy the token (save it securely)

**Configure Secrets**: In the GitHub Repository, navigate to **Settings > Secrets and variables > Actions** and add the following:

* `DOCKER_USERNAME`: Docker Hub ID.
* `DOCKER_PASSWORD`: Personal Access Token (PAT) 
*(generated from the Docker Hub account.)*
* `PERSONAL_ACCESS_TOKEN`: A GitHub PAT with ('repo') scope to allow the runner to push changes to your GitOps repository.

#### Important Note: **If you are cloning or forking this repo make sure to modify the **Image Name**: Update the `env:` section in the `.github/workflows/ci.yaml` file to reflect your own Docker Hub username and repository paths.

## Build & Deployment Strategy

### For this repo the Security Scanning and Build/Deploy workflows are separated into two distinct YAML files.
 Typically in a production context the security scanning tools are fully 
 integrated in order to enforce strict quality gates (no pass, no build). 
 By intentially decoupling these processes, we allow for a full demonstration of the POC while still recognizing and addressing any vulnerabilities if they exist.


## The High-Level CI Build Workflow:

| Step | Action                     | Description                                           |
| :--- | :------------------------- | :---------------------------------------------------- |
| **1.** | WORKFLOW_INITIALIZATION    | ➡️ Monitor branches/paths to trigger automation       |
| **2.** | ESTABLISH_ENV_VARS         | ➡️ Define global constants and repository paths       |
| **3.** | PROVISION_RUNNER_JOB       | ➡️ Provision fresh Ubuntu VM for clean environment    |
| **4.** | EXTRACT_IMAGE_METADATA     | ➡️ Generate unique SHA identity tags for tracking     |
| **5.** | BUILD_IMAGE_AND_PUSH       | ➡️ Package container and upload to registry           |
| **6.** | INTEGRATION_SMOKE_TEST     | ➡️ Verify health endpoint for image functionality     |
| **7.** | GITOPS_REPO_CHECKOUT       | ➡️ Pull deployment repo via secure token for sync     |
| **8.** | UPDATE_DEPLOY_MANIFEST     | ➡️ Inject new Short SHA into Kubernetes YAML files    |
| **9.** | FINAL_COMMIT_AND_PUSH      | ➡️ Sync changes to GitOps to trigger ArgoCD rollout   |

*Note on Deployment: This repository handles the Continuous Integration. Once the image is pushed to Docker Hub and the manifest is updated, the Continuous Deployment is handled by Repo 2 (GitOps). ArgoCD will detect the new image tag and automatically synchronize the Kubernetes cluster state.*

## CI Build & Push YAML


***Workflow Initialization*** 

```yaml
on:
  push:
    branches: [main]
    paths: 
      - 'application/**'
      - 'infrastructure/docker/Dockerfile'
  pull_request:
    branches: [main]
    paths: ['application/**']
```
**Monitors the repository for changes in the 'application' and 'infrastructure' folders to "wake up" the runner. Recognizes a Pull Request for proposed code, while a Push to the 'main' branch initiates the final, official Docker build and security scan**

***Establish Env. Variables*** 

```yaml
env:
  REGISTRY: docker.io
  IMAGE_NAME: nc-fttx-portal
  GITOPS_REPO_PATH: 'nc-fttx-portal-gitops'
  DEPLOYMENT_MANIFEST_PATH: 'manifests/deployment.yaml
```

**This configuration provisions a temporary runner VM to execute the build process exclusively for non-pull request triggering events. It then utilizes a security gate to establish the login for DockerHUB***

***Job Setup*** 

```yaml
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
```

**Establishes a runner VM to test the code on every non pull trigger event request to ensure it builds correctly. It then uses a security gate to establish the login**



***Extract Metadata***

```YAML
- name: Extract metadata
  id: meta
  uses: docker/metadata-action@v5
  with:
    images: ${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}
    tags: |
      type=ref,event=branch
      type=sha,prefix=sha-
      type=raw,value=latest,enable={{is_default_branch}}
```
**Generates a unique identity hash for every build. Useful for auditing code commits.**

***Build Image & Push***

```yaml
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
```
**Builds app image and pushes it to the registry.  Looks for existing layers in cache before building from scratch**
8

***Integration Smoke Test***

```YAML
- name: Test container
  if: github.event_name != 'pull_request'
  run: |
    docker run -d -p 8080:8080 --name test-container ${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}:latest
    sleep 10
    curl -f http://localhost:8080/health || exit 1
    docker stop test-container
    docker rm test-container
```
**Spins up a test container to attempting to reach the health endpoint, to confirm that the application can successfully start and handle requests, preventing a broken image from reaching the registry**

***Updates the Deployment Repo w/ new image***

```yaml
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
```

**Updates the deployment repository with the new image. It only runs if the previous build-and-push job succeeds**


***Update image tag in deployment manifest***

```YAML
- name: Update image tag in deployment manifest
  run: |
    cd ${{ env.GITOPS_REPO_PATH }}
    SHORT_SHA=$(echo "${{ github.sha }}" | cut -c1-7)
    NEW_TAG="${{ secrets.DOCKER_USERNAME }}/${{ env.IMAGE_NAME }}:sha-${SHORT_SHA}"
    
    sed -i 's|image: .*/nc-fttx-portal:.*|image: '"$NEW_TAG"'|g' ${{ env.DEPLOYMENT_MANIFEST_PATH }}
```
**Adds a "Short SHA" tag to the deployment image, to ensure ArgoCD pulls the exact tested version into the cluster.**

***Commit & Push***

```yaml
    - name: Commit and push changes                    
      uses: EndBug/add-and-commit@v9
      with:
        message: 'chore: update application image to sha-${{ github.sha }} [skip ci]' 
        cwd: ${{ env.GITOPS_REPO_PATH }}
        add: '${{ env.DEPLOYMENT_MANIFEST_PATH }}'
        default_author: github_actions
```

***Final stage completes the image updates and pushes them to the deployment repo***

***Continue to the Deployment Repo here: -->***
https://github.com/jaycloud336/nc-fttx-portal-gitops

## ***Important Note:***

**Production Architecture vs. Repo Implementation**

*In a standard enterprise environment, this pipeline would be distributed across multiple environments (Dev, Staging, Prod). For the purpose of this Repo, a simplified Single-Branch strategy is used to demonstrate the immediate feedback loop between code changes and GitOps synchronization.*