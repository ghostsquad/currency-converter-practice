function(USERHOME)

  assert USERHOME != "";

  (import "../config.libsonnet") +
  {
    apiVersion: "k3d.io/v1alpha4",
    kind: "Simple",
    metadata+: {
      name: $.config_.kubernetes.k3d.clusterName,
    },
    servers: 1,
    agents: 0,
    image: "docker.io/rancher/k3s:v%s-k3s1" % [$.config_.kubernetes.expectedVersion],
    kubeAPI+: {
      // TODO spec.json could probably be generated from jsonnet as well
      // We care about this port number in that file...
      hostPort: "53738",
    },
    ports: [
      {
        port: $.config_.kubernetes.k3d.hostPort + ":80",
        nodeFilters: [ "loadbalancer" ],
      },
    ],
    registries: {
      create: {
        name: $.config_.kubernetes.k3d.registry.name,
        host: "0.0.0.0",
        hostPort: std.toString($.config_.kubernetes.k3d.registry.hostPort),
        volumes: [
          USERHOME + "/" + $.config_.kubernetes.k3d.registry.localDataPathRelative + ":/var/lib/registry"
        ]
      }
    },
  }