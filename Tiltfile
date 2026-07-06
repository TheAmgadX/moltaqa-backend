#  ---------- configurations ----------
k8s_yaml("k8s/config.yml")
k8s_yaml("k8s/secrets.yml")

# ---------- api-gateway start ----------

k8s_yaml("services/api-gateway/deployments/api-gateway.yml")
docker_build(
    "api-gateway",
    ".",
    dockerfile="services/api-gateway/deployments/Dockerfile",
)

k8s_resource("api-gateway", port_forwards='8080:8080')

# ---------- api-gateway end ----------
