# Sandbox Verification Guide (No Git Required)

See the application in action immediately without cloning a repository or setting up a local development environment, use Killer.sh Ubuntu or the Sandbox of your choice.

Quick and easy: https://killercoda.com/playgrounds/scenario/ubuntu

### **Pull and Run the App Image**
By pulling directly from Docker Hub, you bypass the build phase to verify the application's final state:

**Pull the public image from Docker Hub**
docker pull jaycloud336/nc-fttx-portal:latest

**Run the container in detached mode, mapping port 8080**
docker run -d -p 8080:8080 --name prod-baseline jaycloud336/nc-fttx-portal:latest

**List running containers**
docker ps

**Internal verification via curl**
curl -I http://localhost:8080

Go to he "Traffic Port Accessor"
Access HTTP services which run in your environment
Located in the top right corner of the killercoda GUI 
Choose Common port "8080" and Access the generated URL to view the live application in action!





## Build from the Source: Local Container Testing
Now that the application has been verified via the baseline image, proceed with building the image directly from the source code to validate the Multi-stage Dockerfile configuration.

Clone the repository to your local env , your sandbox env or dev continer and navigate to the application directory to review the Go source files:

**Clone the project repository**
`git clone https://github.com/jaycloud336/nc-fttx-portal)`

**Navigating to the application directory**
`cd nc-fttx-portal/application`

### Optional: Before building, you may run the app natively
**To verify Go dependencies first run check if you have GO installed: (If not, proceed to container build)**
`go version`
**Then run:**
`go run main.go`
*Note: This app is designed to run on port 8080*
*After running this command, access the app via the follwing url: "http://locahost:8080" in your browser*

### Container Build:
**Build the image using the specified infrastructure path**
`docker build -t nc-fttx-portal:local -f ./infrastructure/docker/Dockerfile ./application`

**Check the image**
`docker images`

**Run the container in detached mode, mapping port 8080**
`docker run -d -p 8080:8080 --name test-portal nc-fttx-portal:latest`

**Wait for container to build then check the container**
`docker ps`

**Test access with curl commands:**
`curl -f http://localhost:8080/health`
`curl -f http://localhost:8080`

**Access app via your browser:**
`Access app via browser url: http://localhost:8080`

**Access app via your killercoda sandbox instance:**
If you are using a  killercoda instance as sandbox, go to he "Traffic Port Accessor" Access HTTP services which run in your environment Located in the top right corner of the killercoda GUI Choose Common port "8080" and Access the generated URL to view the live application in action!


**Clean up the container (local env)**
`docker stop test-portal`
`docker rm test-portal`

**Optionally, clean up image (local env)**
`docker rmi nc-fttx-portal`




