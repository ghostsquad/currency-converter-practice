function(tag)

  (import "../base.libsonnet") + {
    config_+: {
      image+: {
        full: std.join(":", [self.imageNoTag, tag]),
      },
    },
  }