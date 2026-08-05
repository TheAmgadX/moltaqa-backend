# Moltaqa Backend

This is the backend repository for the Moltaqa project. It is built using Go, gRPC, PostgreSQL, Redis, Kafka, and deployed locally via Kubernetes and Tilt.

## Prerequisites & Setup

### Windows

1. **Install Go**: Download and install the latest version of Go from [golang.org](https://go.dev/dl/).
2. **Install Docker Desktop**: Download and install from [docker.com](https://www.docker.com/products/docker-desktop).
3. **Enable Kubernetes**: Open Docker Desktop settings, navigate to the **Kubernetes** tab, and check "Enable Kubernetes". Apply the changes and wait for the cluster to start.
4. **Install Tilt**: Download and install Tilt. For Windows, you can use a package manager like Scoop (`scoop install tilt`) or download the binary directly from [tilt.dev](https://tilt.dev/).
5. **Get Secrets**: Obtain the necessary secret files from the project administrator and place them in their respective locations (or create your own based on sample configurations). The required ignored secret files are:
   - `.env` (Project root)
   - `secrets.yml` (Project root)
   - `services/user-service/deployments/k8s/postgres_secrets.yml`
6. **Run the Project**: Open your terminal (PowerShell or Command Prompt) in the project root and run:
   ```bash
   tilt up
   ```

### Linux

1. **Install Go**: Use your distribution's package manager or download from [golang.org](https://go.dev/dl/).
   ```bash
   sudo apt-get update
   sudo apt-get install golang-go
   ```
2. **Install Docker**: Install Docker Engine.
   ```bash
   sudo apt-get install docker.io
   sudo systemctl enable --now docker
   sudo usermod -aG docker $USER
   ```
   *(Note: You may need to log out and log back in for the group changes to take effect.)*
3. **Install Local Kubernetes Cluster**: You can use `Minikube`, `Kind`, or `k3d`. For example, using `Kind`:
   ```bash
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind
   kind create cluster
   ```
4. **Install Tilt**: Run the Tilt installation script:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
   ```
5. **Get Secrets**: Obtain the necessary secret files from the project administrator and place them in their respective locations (or create your own based on sample configurations). The required ignored secret files are:
   - `.env` (Project root)
   - `secrets.yml` (Project root)
   - `services/user-service/deployments/k8s/postgres_secrets.yml`
6. **Run the Project**: Open your terminal in the project root and run:
   ```bash
   tilt up
   ```

## Additional Information

- **Tilt Dashboard**: To view the Tilt UI, press `Space` in the terminal after running `tilt up` or navigate to `http://localhost:10350` in your browser.
- **Hot Reloading**: The `tilt up` command will automatically build the Docker images, deploy the Kubernetes manifests, and set up port forwarding. It also monitors for file changes and live-updates your containers.
