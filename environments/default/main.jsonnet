function(tag)

  (import "../base.libsonnet") + {
    config_+: {
      image+: {
        full: std.join(":", [self.imageNoTag, tag]),
        imageNoTag: self.k3dRemoteNoTag,
      },
    },

    local k = (import "k.libsonnet")($.config_.kubernetes.versionShort),
    local networkingV1 = k.networking.v1,

    // TODO config instead of local?
    local app = "app",
    local namespace = $.config_.project.name,
    local labels = $.config_.kubernetes.labels,

    ingress:
      networkingV1.ingress.new(app) +
      networkingV1.ingress.metadata.withNamespace(namespace) +
      networkingV1.ingress.metadata.withLabels(labels) +
      networkingV1.ingress.spec.withRules([
        // TODO find/make rule function
        {
          http: {
            paths: [
              {
                path: "/",
                pathType: "Prefix",
                backend: {
                  service: {
                    name: $.service.metadata.name,
                    port: {
                      name: $.config_.app.ports.http.name,
                    },
                  }
                },
              },
            ],
          },
        },
      ])
    ,
  }
