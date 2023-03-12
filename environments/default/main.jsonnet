(import "../base.libsonnet") + {
  config_+: {
    image+: {
      imageNoTag: self.k3dRemoteNoTag,
    },
  },
}
