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

env = {
  'PKG_CONFIG_PATH': meson.current_build_dir() + '/meson-private',
  'CPATH': meson.current_source_dir() + '/subprojects/ffmpeg' + ':' + meson.current_build_dir() + '/subprojects/ffmpeg',
}
GOOS = meson.get_external_property('GOOS', host_machine.system())
GOARCH = meson.get_external_property('GOARCH', host_machine.cpu())
if GOARCH == 'x86_64'
  GOARCH = 'amd64'
endif
if GOARCH == 'aarch64'
  GOARCH = 'arm64'
endif
if meson.is_cross_build()
  env += {
    'GOOS': GOOS,
    'GOARCH': GOARCH,
    'CGO_ENABLED': '1',
    'CC': meson.get_compiler('c', native: false).cmd_array()[0],
  }
endif

c2pa_go_opts = {}
triple = meson.get_external_property('RUST_TRIPLE', '')
if triple != ''
  c2pa_go_opts += {'RUST_TRIPLE': triple}
endif
c2pa_go_proj = subproject(
  'c2pa_go',
  default_options: c2pa_go_opts,
)
c2pa_go_dep = c2pa_go_proj.get_variable('c2pa_go_dep')

mistserver_proj = subproject(
  'mistserver',
  default_options: {
    'APPNAME': 'MistServer',
  },
)
mistserver = mistserver_proj.get_variable('libmistserver_dep')

ffmpeg_proj = subproject(
  'ffmpeg',
  # vvc seems to have problems building as a submodule?
  default_options: {
    'vvc_decoder': 'disabled',
    'vvc_demuxer': 'disabled',
    'vvc_metadata_bsf': 'disabled',
    'vvc_mp4toannexb_bsf': 'disabled',
    'vvc_muxer': 'disabled',
    'vvc_parser': 'disabled',
    'signature_cuda_filter': 'enabled',
    'signature_filter': 'enabled',
    'gpl': 'enabled',
  },
)
avformat = ffmpeg_proj.get_variable('libavformat_dep')
avfilter = ffmpeg_proj.get_variable('libavfilter_dep')

libaquareum = custom_target(
  'libaquareum',
  input: [aquareum_go, flags],
  output: ['aquareum.a', 'aquareum.h'],
  # depends: [c2pa_go_custom],
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

aquareum_deps = []
if host_machine.system() == 'darwin'
  aquareum_deps += [dependency('appleframeworks', modules: ['CoreFoundation', 'Security'])]
endif
aquareum = executable(
  'aquareum',
  ['cmd/aquareum/aquareum.c', libaquareum],
  dependencies: [avformat, avfilter, mistserver, aquareum_deps],
  link_with: [c2pa_go_dep],
)

prog_tar = find_program('tar')
# archive_name = aquareum-$(VERSION)-$$GOOS-$$GOARCH.tar.gz
archive_name = 'aquareum-' + version + '-' + GOOS + '-' + GOARCH + '.tar.gz'
custom_target(
  'archive',
  input: [aquareum],
  output: [archive_name],
  command: [prog_tar, '-czvf', '../bin/@OUTPUT0@', '@INPUT@'],
  install: true,
  install_dir: './bin',
  build_by_default: false,
)