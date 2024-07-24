project('aquareum', 'c', 'cpp', default_options: ['cpp_std=c++98'])

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

mistserver_proj = subproject('mistserver', default_options: {
  'APPNAME': 'MistServer',
})
mistserver = mistserver_proj.get_variable('libmistserver_dep')

executable('aquareum', ['cmd/aquareum/aquareum.c', libaquareum], dependencies: [mistserver])
