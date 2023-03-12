{
  config_+:: {
    local c = self,
    project+: {
      name: 'currency-converter-practice',
      owner: 'ghostsquad',

      // this is a duplication of what's in layers/taskfile/go.libsonnet
      // but tanka (tk) doesn't see these fields, because it's supposed to technically just be for taskfile
      // TODO need to do a better job of separating config so it can be unified and used everywhere
      repoShort: std.join("/", [self.owner, self.name]),
      repoLong: std.join("/", ["github.com", self.repoShort]),
    },
    image+: {
      full: error "image.full isn't available without a tag, read at runtime",
      imageNoTag: self.remoteNameNoTag,
      remoteNameNoTag: std.join("/", ["docker.io", c.project.repoShort]),
      k3dRemoteNoTag: std.join("/", [c.kubernetes.k3d.registry, c.project.repoShort])
    },
    app+: {
      ports+: {
        // TODO plumb this into port ENV configuration, Dockerfile, Tests, etc...
        http+: {
          name: "http",
          port: 8080,
        },
      },
    },
    go+: {
      expectedVersion: '1.20.2',
    },
    kubernetes+: {
      // TODO tie this to managing the .tool-versions file as well, like we do with `go`
      expectedVersion: '1.26.2',
      // NOTE: with array slicing:
      // The result includes the start index, but excludes the end index.
      // https://jsonnet.org/learning/tutorial.html (ctrl+f slices)
      // https://www.w3schools.com/python/numpy/numpy_array_slicing.asp
      versionShort: std.join('.', std.split(self.expectedVersion, '.')[0:2]),
      k3d+: {
        registryNameShort: "registry-default",
        registryName: std.join("-", ["k3d", self.registryNameShort]),
        registry: std.join(":", [self.registryName, std.toString(self.registryPort)]),
        registryPort: 5100,
        clusterName: 'k3s-default',
      },
      labels: {
        app: c.project.name,
      },
    },
    GetEnvironmentRootRelPath(name):: './environments/' + name,
  },
}