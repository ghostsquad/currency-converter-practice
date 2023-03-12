local kubeLibs = {
  "1.21": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.21/main.libsonnet'),
  "1.22": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.22/main.libsonnet'),
  "1.23": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.23/main.libsonnet'),
  "1.24": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.24/main.libsonnet'),
  "1.25": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.25/main.libsonnet'),
  "1.26": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.26/main.libsonnet'),
};

function(version) kubeLibs[version]
