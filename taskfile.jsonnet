local taskfile = import "github.com/ghostsquad/practice-layers/taskfile/taskfile.libsonnet";

// TODO tanka doesn't see config items set in these layers! :(
// Error: evaluating jsonnet: RUNTIME ERROR: Field does not exist: repoShort
taskfile +
(import "github.com/ghostsquad/practice-layers/taskfile/jb.libsonnet") +
(import "github.com/ghostsquad/practice-layers/taskfile/go.libsonnet") +
(import "./config.libsonnet") +
{
  env+: {
    // TODO contribute this upstream
    GOFLAGS: "-mod=mod",
  },
  vars+: {
    USERHOME: '{{env "HOME"}}',
    EXPECTED_GO_VERSION: $.config_.go.expectedVersion,
    K3D_APP_IMAGE: std.join("/", [$.config_.kubernetes.k3d.registry.nameAndHostPort, $.config_.project.repoShort]) + ":{{.GIT_COMMIT}}"
  },
  tasks+: {
    # TODO Var order matters, but since this file is dynamically generated from jsonnet, we lose ordering
    # Regenerating based on this file will result in problems.
    # !!! Don't forget to fix the ordering in Taskfile.yml in the interm
    # https://github.com/go-task/task/issues/1051
    "taskfile:gen"+: {
      deps+: [$.tasks["jb:install"].name_],
    },
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
    },
    // TODO contribute this upstream
    passthru: taskfile.Task("passthru")
      .WithCmds("{{.CLI_ARGS}}")
    ,
    // TODO contribute this upstream
    tk: taskfile.Task("tk")
      .WithCmds("tk {{.CLI_ARGS}}")
      // TODO Taskfile is becoming a burden... might be time to pickup fn-go again
      // {{.CLI_ARGS}} are passed to jb:install as well, which is a problem
      //.WithDeps($.tasks["jb:install"].name_)
    ,
    "tk:show": taskfile.Task("tk:show")
      .WithCmds("tk show --tla-str tag={{.GIT_COMMIT}} {{.CLI_ARGS}}")
    ,
    "tk:apply": taskfile.Task("tk:apply")
      .WithCmds("tk apply --tla-str tag={{.GIT_COMMIT}} {{.CLI_ARGS}}")
    ,
    "publish:k3d": taskfile.Task("publish:k3d")
      .WithCmds(
        taskfile.CmdTask($.tasks.build.name_)
          .WithVars({
            PUSH_IMAGE: '{{.K3D_APP_IMAGE}}',
            BUILD_ARGS: |||
              --platform linux/amd64 \
              --push
            |||,
          })
      )
    ,
    // We are overriding upstream test:integration (for now)
    'test:integration': taskfile.Task('test:integration')
      .WithCmds(
        |||
          APP_IMAGE='{{.APP_IMAGE}}' \
          docker-compose \
            --file {{.DOCKERFILE}} \
              up \
              --exit-code-from test \
              --abort-on-container-exit \
            ;
        |||
      )
      .WithVars({
        DOCKERFILE: '{{.DOCKERFILE | default "docker-compose.tests.integration.yml"}}',
      })
      .WithDeps([
        taskfile.CmdTask($.tasks.build.name_)
          .WithVars({
            BUILD_ARGS: '--output=type=docker',
          })
        ,
      ]) + {
        env+: {
          SUBJECT_HOSTPORT: $.config_.app.ports.http.port,
        },
      }
    ,
    'test:integration:k3d': taskfile.Task('test:integration:k3d')
      .WithCmds(
        |||
          ./hack/test.sh 'localhost' '%d'
        ||| % [$.config_.kubernetes.k3d.hostPort]
      )
    ,
    // TODO DRY up the variables in this task
    "k3d:config:gen": taskfile.Task("k3d:config:gen")
      .WithCmds([
        |||
          mkdir -p '{{.USERHOME}}/%s'
        ||| % [$.config_.kubernetes.k3d.registry.localDataPathRelative],
        |||
          jsonnet --tla-str USERHOME='{{.USERHOME}}' ./environments/default/k3d.jsonnet | dasel -f - -r json -w yaml --pretty > ./environments/default/k3d.yml
        |||,
        |||
          echo "# Code generated by task taskfile:gen; DO NOT EDIT." \
            | cat - ./environments/default/k3d.yml \
            | sponge ./environments/default/k3d.yml
        |||
      ])
      .WithDeps($.tasks["jb:install"].name_)
    ,
    // TODO add status to skip create if the cluster already exists
    // TODO add a start if cluster is created by not started
    "k3d:up": taskfile.Task("k3d:create")
      .WithCmds([
        "k3d cluster create --config ./environments/default/k3d.yml",
        |||
          echo "edit /etc/hosts with 127.0.0.1 %s"
        ||| % [$.config_.kubernetes.k3d.registry.name],
      ])
      .WithDeps([
        $.tasks["k3d:config:gen"].name_,
      ])
    ,
  }
}
