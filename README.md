## NC FTTX Portal - Automated CI Pipeline & Artifact Engineering

### Continous Integration Repository
This repository manages the Continuous Integration (CI) for the NC FTTX Portal, a Go-based telecom infrastructure application. It is engineered to transform raw source code into K8-ready container images. By isolating the CI process here, we ensure that every artifact is verified, tested, and scanned before it ever reaches the deployment phase.

### Continous Deployment Repo
*There is a second repository associated with this project as well.
It can be located here:*

`https://github.com/jaycloud336/nc-fttx-portal-gitops`

*The `nc-fttx-portal-gitops` repository handles the Continuous Deployment process. Once the image is pushed to Docker Hub, ArgoCD will detect the new image tag and automatically synchronize the Kubernetes cluster state. All of tke K8 & ArgoCD manifest files are located
in the Continuous Deployment repository.*

![NC FTTX Portal CI Workflow](./Overview.png)

### Major DevOps Components for artifact management:

**Artifact Hardening:** Utilizes multi-stage builds to reduce image size, create minimal Go binaries and reduce overall attack surface.

**Security Scanning:** Integrated container scanning to detect vulnerabilities within image layers prior to registry storage.

**CI-to-CD Handshake:** Automated triggers that update a second CD repository once a new verified image is pushed.

### CI Workflow: 

**Developer (Push) ➡️ GitHub Actions (CI) ➡️ Multi-Stage Build ➡️ Security Scan ➡️ Docker Hub (Registry) ➡️ GitOps Trigger (Update Tags)**

***Source Code Validation:** Automated testing of the Go application to ensure functional integrity.*

***Optimized Build:** Execution of multi-stage Docker builds to package the application efficiently.*

***Security Audit:** Vulnerability scanning of the resulting image to meet production security requirements.*

***Registry Promotion:** Pushing the verified image to Docker Hub with unique commit-SHA tagging.*

***GitOps Trigger:** Updating the deployment manifests in the companion CD repository.*


### 3. Get Started - Clone or Fork Project Repo and follow the links:

***Verfication Guide*** (preview the working application)

***Step by step CI Pipeline Build*** 
(Walk through the pipeline build process)

***Step by step Security Scan Pipeline Build***
(Walk through the pipeline build process)

***Build from scratch implementation guide***
(Create your own repo & directory structure from scratch)
