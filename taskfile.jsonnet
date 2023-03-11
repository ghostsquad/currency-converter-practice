local taskfile = import "github.com/ghostsquad/practice-layers/taskfile/taskfile.libsonnet";

taskfile +
(import "github.com/ghostsquad/practice-layers/taskfile/jb.libsonnet") +
(import "github.com/ghostsquad/practice-layers/taskfile/go.libsonnet") +
{
  config_+:: {
    project+: {
      name: "currency-converter-practice",
      owner: "ghostsquad",
    }
  },
  # TODO Var order matters, but since this file is dynamically generated from jsonnet, we lose ordering
  # Regenerating based on this file will result in problems.
  # !!! Don't forget to fix the ordering in Taskfile.yml in the interm
  # https://github.com/go-task/task/issues/1051
  vars+: {
    EXPECTED_GO_VERSION: "1.20.2",
  },
  tasks+: {
    run+: {
      cmds: [
        // TODO might be better for upstream not to assume `./...` is the path
        // and we can use a hidden/task variable instead
        "go run ./cmd/app/..."
      ]
    },
    build+: {
      deps+: [
        // TODO contribute this upstream
        taskfile.CmdTask("go:version:verify")
      ]
    }
  }
}
