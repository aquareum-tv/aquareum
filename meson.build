project(
  'aquareum',
  'c',
  'cpp',
  default_options: ['cpp_std=c++98', 'default_library=static'],
)

fs = import('fs')

prog_go = find_program('go')

flagbuilder = custom_target(
  'flagbuilder',
  input: ['pkg/config/git/git.go'],
  output: ['flagbuilder'],
  command: [prog_go, 'build', '-o', '@OUTPUT0@', '@INPUT@'],
  build_always_stale: true,
)

flags = custom_target(
  'flags',
  input: [],
  output: ['build-flags.go'],
  command: [flagbuilder, '-o', '@OUTPUT0@'],
  build_always_stale: true,
)

version_cmd = run_command(prog_go, 'run', 'pkg/config/git/git.go', '-v', check: true)
version = version_cmd.stdout().strip()

aquareum_go = custom_target(
  'aquareum.go',
  input: 'cmd/libaquareum/aquareum.go',
  output: 'aquareum.go',
  command: ['cp', '@INPUT@', '@OUTPUT@'],
)

env = {}
GOOS = meson.get_external_property('GOOS', host_machine.system())
GOARCH = meson.get_external_property('GOARCH', host_machine.cpu())
if GOARCH == 'x86_64'
  GOARCH = 'amd64'
endif
if GOARCH == 'aarch64'
  GOARCH = 'arm64'
endif
if meson.is_cross_build()
  env = {
    'GOOS': GOOS,
    'GOARCH': GOARCH,
    'CGO_ENABLED': '1',
    'CC': meson.get_compiler('c', native: false).cmd_array()[0],
  }
endif
libaquareum = custom_target(
  'libaquareum',
  input: [aquareum_go, flags],
  output: ['aquareum.a', 'aquareum.h'],
  command: [
    prog_go,
    'build',
    '-o', '@OUTPUT0@',
    '-buildmode', 'c-archive',
    '-tags', 'netgo',
    '@INPUT@',
  ],
  build_always_stale: true,
  env: env,
)

mistserver_proj = subproject(
  'mistserver',
  default_options: {
    'APPNAME': 'MistServer',
  },
)
mistserver = mistserver_proj.get_variable('libmistserver_dep')

aquareum_deps = []
if host_machine.system() == 'darwin'
  aquareum_deps += [dependency('appleframeworks', modules: ['CoreFoundation', 'Security'])]
endif
aquareum = executable(
  'aquareum',
  ['cmd/aquareum/aquareum.c', libaquareum],
  dependencies: [mistserver, aquareum_deps],
)

prog_tar = find_program('tar')
# archive_name = aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz
archive_name = 'aquareum-' + version + '-' + GOOS + '-' + GOARCH + '.tar.gz'
archive = custom_target(
  'archive',
  input: [aquareum],
  output: [archive_name],
  command: [prog_tar, '-czvf', '../bin/@OUTPUT0@', '@INPUT@'],
  install: true,
  install_dir: './bin',
  build_by_default: false,
)