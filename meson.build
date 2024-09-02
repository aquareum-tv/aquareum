project(
  'aquareum',
  'c',
  'cpp',
  default_options: {
    'cpp_std': 'c++11',
    'default_library': 'static',
    'auto_features': 'disabled',
    'force_fallback_for': 'glib-2.0,gobject-2.0,gio-2.0,gio-unix-2.0',
    'gst-plugins-base:audioresample': 'enabled',
    'gst-plugins-base:opus': 'enabled',
    'gst-plugins-good:matroska': 'enabled',
    'gst-plugins-bad:fdkaac': 'enabled',
    'buildtype': 'debugoptimized',
  },
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

# todo: teach meson to automatically populate this CPATH thing for go
env = {
  'PKG_CONFIG_PATH': meson.current_build_dir() + '/meson-private',
  'CPATH': meson.current_source_dir()  + '/subprojects/FFmpeg' 
    + ':' + meson.current_build_dir()  + '/subprojects/FFmpeg'
    + ':' + meson.current_source_dir() + '/subprojects/gstreamer-full/subprojects/gstreamer'
    + ':' + meson.current_build_dir()  + '/subprojects/gstreamer-full/subprojects/gstreamer'
    + ':' + meson.current_source_dir() + '/subprojects/gstreamer-full/subprojects/gstreamer/libs'
    + ':' + meson.current_build_dir()  + '/subprojects/gstreamer-full/subprojects/gstreamer/libs'
    + ':' + meson.current_source_dir() + '/subprojects/gstreamer-full/subprojects/gstreamer'
    + ':' + meson.current_source_dir() + '/subprojects/glib-2.78.3'
    + ':' + meson.current_build_dir()  + '/subprojects/glib-2.78.3'
    + ':' + meson.current_source_dir() + '/subprojects/glib-2.78.3/glib'
    + ':' + meson.current_build_dir()  + '/subprojects/glib-2.78.3/glib'
    + ':' + meson.current_source_dir() + '/subprojects/glib-2.78.3/gmodule'
    + ':' + meson.current_build_dir()  + '/subprojects/glib-2.78.3/gmodule'
    + ':' + meson.current_source_dir() + '/subprojects/proxy-libintl'
    + ':' + meson.current_build_dir()  + '/subprojects/proxy-libintl'
    ,
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

aquareum_deps = []
if host_machine.system() == 'linux'
  mistserver_proj = subproject(
    'mistserver',
    default_options: {
      'APPNAME': 'MistServer',
    },
  )
  mistserver = mistserver_proj.get_variable('libmistserver_dep')
  aquareum_deps += mistserver
endif

ffmpeg_proj = subproject(
  'FFmpeg',
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

gstreamer_proj = subproject(
  'gstreamer-full',
  default_options: {
    'gst-full': 'enabled',
    'gst-full-plugins': 'libgstaudioresample.a;libgstmatroska.a;libgstfdkaac.a;libgstopus.a',
    'gst-full-libraries': 'gstreamer-controller-1.0',
    'gst-full-target-type': 'static_library',
    'bad': 'enabled',
    'ugly': 'disabled',
    'tls': 'disabled',
    'gpl': 'enabled',
    'force_fallback_for': 'glib,gio',
    'gst-full-elements': 'coreelements:fdsrc,fdsink',
    # 'gst-full-typefind-functions': 'typefindfunctions:wav,flv:'
    # 'gst-full-device-providers': 'alsa:alsadeviceprovider:'
    # 'gst-full-dynamic-types': 'pbtypes:video_multiview_flagset:'
  },
)

gst_full_dep = gstreamer_proj.get_variable('gst_full_dep')

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

if host_machine.system() == 'darwin'
  aquareum_deps += [dependency('appleframeworks', modules: ['CoreFoundation', 'Security'])]
endif
if host_machine.system() == 'windows'
  add_project_link_arguments([
    '-lntdll',
    '-luserenv',
    '-lcrypt32',
  ], language : [ 'c', 'cpp' ])
endif
gio = dependency('gio-2.0')
controller = dependency('gstreamer-controller-1.0')
aquareum = executable(
  'aquareum',
  ['cmd/aquareum/aquareum.c', libaquareum],
  dependencies: [avformat, avfilter, gst_full_dep, controller, gio, aquareum_deps],
  link_with: [c2pa_go_dep],
)

if not meson.is_cross_build()
  test('go-tests', prog_go,
    depends: [aquareum],
    env: env,
    verbose: true,
    timeout: 0,
    args : [
    'test',
    '@0@/pkg/...'.format(meson.current_source_dir()),
    '@0@/cmd/...'.format(meson.current_source_dir()),
  ]
)
endif

if host_machine.system() == 'windows'
  prog_zip = find_program('zip')
  archive_name = 'aquareum-' + version + '-' + GOOS + '-' + GOARCH + '.zip'
  custom_target(
    'archive',
    input: [aquareum],
    output: [archive_name],
    command: [prog_zip, '../bin/@OUTPUT0@', '@INPUT@'],
    install: true,
    install_dir: './bin',
    build_by_default: false,
  )
else
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
endif