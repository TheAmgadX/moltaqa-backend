#  ---------- configurations ----------
k8s_yaml("k8s/config.yml")
k8s_yaml("k8s/secrets.yml")

# ---------- kafka infrastructure start ----------

k8s_yaml("k8s/kafka/kafka-nodepool.yml")
k8s_yaml("k8s/kafka/kafka.yml")
k8s_yaml("k8s/kafka/kafka-admin-user.yml")
k8s_yaml("k8s/kafka/config.yml")
k8s_yaml("k8s/kafka/kafka-ui.yml")

k8s_resource("kafka-ui", port_forwards='8090:8080')

# ---------- kafka infrastructure end ----------

# ---------- api-gateway start ----------

k8s_yaml("services/api-gateway/deployments/api-gateway.yml")
docker_build(
    "api-gateway",
    ".",
    dockerfile="services/api-gateway/deployments/Dockerfile",
)

k8s_resource("api-gateway", port_forwards='8080:8080')

# ---------- api-gateway end ----------

# ---------- user-service start ----------

k8s_yaml("services/user-service/deployments/k8s/user-service.yml")
docker_build(
    "user-service",
    ".",
    dockerfile="services/user-service/deployments/docker/Dockerfile",
)

k8s_resource("user-service", port_forwards='50051:50051')

# ---------- user-service end ----------


# ---------- kafka-bootstrap start ----------

k8s_yaml("services/kafka-bootstrap/deployments/k8s/job.yml")
docker_build(
    "kafka-bootstrap",
    ".",
    dockerfile="services/kafka-bootstrap/deployments/docker/Dockerfile",
)

k8s_resource("kafka-bootstrap")

# ---------- kafka-bootstrap end ----------
