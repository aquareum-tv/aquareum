project('mistserver', 'c')

prog_go = find_program('go')

flagbuilder = custom_target('flagbuilder',
  input: ['pkg/config/git/git.go'],
  output: ['flagbuilder'],
  command: [prog_go, 'build', '-o', '@OUTPUT0@', '@INPUT@'],
)

flags = custom_target('flags',
  input: [],
  output: ['build-flags.go'],
  command: [flagbuilder, '-o', 'build-flags.go'],
  build_always_stale: true,
)

aquareum_go = custom_target('aquareum.go',
  input : 'cmd/libaquareum/aquareum.go',
  output :  'aquareum.go',
  command : ['cp', '@INPUT@', '@OUTPUT@'],
)

libaquareum = custom_target('libaquareum',
  input: [aquareum_go, flags],
  output: ['aquareum.a', 'aquareum.h'],
  command: [prog_go, 'build', '-o', '@OUTPUT0@', '-buildmode', 'c-archive', '@INPUT@'],
  build_always_stale: true,
)

# uuid = custom_target('uuid',
#   input: ['pkg/config/uuid/uuid.go'],
#   output: ['uuid'],
#   command: [prog_go, 'build', '-o', '@OUTPUT0@', '@INPUT@']
# )

# uuid_txt = custom_target('uuid_txt',
#   output: ['uuid.txt'],
#   command: [uuid, '-o', '@OUTPUT0@'],
#   build_always_stale: true,
# )

executable('aquareum', ['cmd/aquareum/aquareum.c', libaquareum])
