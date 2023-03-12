(import "../config.libsonnet") +
{
  apiVersion: "k3d.io/v1alpha4",
  kind: "Simple",
  metadata: {
    name: $.config_.kubernetes.k3d.clusterName,
  },
  servers: 1,
  agents: 0,
  image: "docker.io/rancher/k3s:v%s-k3s1" % [$.config_.kubernetes.expectedVersion],
}