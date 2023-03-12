local lm = import "github.com/cakehappens/lonely-mountain/main.libsonnet";

// it's in the vendor directory, which is part of the path
(import "../config.libsonnet") +
{
  local k = (import "k.libsonnet")($.config_.kubernetes.versionShort),
  local appsV1 = k.apps.v1,
  local coreV1 = k.core.v1,

  // TODO config instead of local?
  local app = "app",

  deployment:
    appsV1.deployment.new(app) +
    appsV1.deployment.metadata.withLabels($.config_.kubernetes.labels) +
    appsV1.deployment.metadata.withNamespace($.config_.project.name) +
    appsV1.deployment.spec.template.metadata.withLabels($.config_.kubernetes.labels) +
    appsV1.deployment.spec.template.spec.withContainers(
      coreV1.container.withName(app) +
      coreV1.container.withPorts(lm.obj.valuesPruned($.config_.app.ports))
    )
  ,
  service:
    coreV1.service.new(
      app,
      // TODO assert that whatever value is used here is a subset of deployment.spec.template.metadata.labels
      {matchLabels: $.deployment.spec.template.metadata.labels},
      lm.obj.valuesPruned($.config_.app.ports)
    ) +
    coreV1.service.metadata.withNamespace($.config_.project.name)
  ,
}