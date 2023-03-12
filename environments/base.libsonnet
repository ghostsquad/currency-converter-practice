local lm = import "github.com/cakehappens/lonely-mountain/main.libsonnet";

// it's in the vendor directory, which is part of the path
(import "../config.libsonnet") +
{
  local k = (import "k.libsonnet")($.config_.kubernetes.versionShort),
  local appsV1 = k.apps.v1,
  local coreV1 = k.core.v1,

  // TODO config instead of local?
  local app = "app",
  local namespace = $.config_.project.name,
  local labels = $.config_.kubernetes.labels,

  namespace: coreV1.namespace.new(namespace),

  deployment:
    appsV1.deployment.new(app) +
    appsV1.deployment.metadata.withLabels(labels) +
    appsV1.deployment.metadata.withNamespace(namespace) +
    appsV1.deployment.spec.selector.withMatchLabels(labels) +
    appsV1.deployment.spec.template.metadata.withLabels(labels) +
    appsV1.deployment.spec.template.spec.withContainers(
      coreV1.container.withName(app) +
      coreV1.container.withPorts([
        coreV1.toContainerPort(p),
        for p in lm.obj.valuesPruned($.config_.app.ports)
      ]) +
      coreV1.container.withImage($.config_.image.full)
    )
  ,
  service:
    coreV1.service.new(
      app,
      // TODO assert that whatever value is used here is a subset of deployment.spec.template.metadata.labels
      $.deployment.spec.template.metadata.labels,
      [
        coreV1.toServicePort(p),
        for p in lm.obj.valuesPruned($.config_.app.ports)
      ]
    ) +
    coreV1.service.metadata.withNamespace(namespace)
  ,
}