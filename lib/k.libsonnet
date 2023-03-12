local lm = import "github.com/cakehappens/lonely-mountain/main.libsonnet";

local kubeLibs = {
  "1.21": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.21/main.libsonnet'),
  "1.22": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.22/main.libsonnet'),
  "1.23": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.23/main.libsonnet'),
  "1.24": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.24/main.libsonnet'),
  "1.25": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.25/main.libsonnet'),
  "1.26": (import 'github.com/jsonnet-libs/k8s-libsonnet/1.26/main.libsonnet'),
};

function(version) kubeLibs[version] + {
  core+: {
    v1+: {
      // TODO there are other fields that we probably should support
      // https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#containerport-v1-core
      toContainerPort(obj):: {
        name: obj.name,
        containerPort: lm.obj.lookup(obj, "port", lm.obj.lookup(obj, "containerPort", error "port|containerPort not found"))
      },
      toServicePort(obj):: {
        name: obj.name,
        port: lm.obj.lookup(obj, "port", lm.obj.lookup(obj, "containerPort", error "port|containerPort not found"))
      }
    }
  }
}
