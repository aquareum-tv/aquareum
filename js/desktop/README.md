# Aquareum Desktop

## `yarn run start`

Boot up Aquareum Desktop for development. Two environment variable options:

- `AQD_SKIP_NODE`: (default `false`): Skip booting up an Aquareum node from
  within the app so you can just test the app against the node you're already
  running locally.

- `AQD_NODE_FRONTEND`: (default `false`): Use frontend bundled with the Aquareum
  node instead of the dev server on `http://localhost:8081`.
