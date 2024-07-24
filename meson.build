project('mistserver', 'c')

prog_go = find_program('go')
libaquareum = custom_target('libaquareum',
  input: ['cmd/aquareum/aquareum.go'],
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
