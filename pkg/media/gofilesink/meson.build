gofilesink = custom_target(
  'gofilesink',
  input: ['gofilesink.go', 'zzgenerated_plugin.go'],
  output: ['libgstgofilesink.a'],
  command: [
    prog_go,
    'build',
    '-o', '@OUTPUT0@',
    '-buildmode', 'c-archive',
    '-tags', 'netgo',
    '@INPUT@',
  ],
  env: env,
)