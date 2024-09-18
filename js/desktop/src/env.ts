export default function getEnv() {
  const isDev = process.env["WEBPACK_SERVE"] === "true";
  if (!isDev) {
    return {
      isDev: false,
      skipNode: false,
      nodeFrontend: true,
    };
  }
  return {
    isDev: process.env["WEBPACK_SERVE"] === "true",
    skipNode: process.env["AQD_SKIP_NODE"] === "true",
    nodeFrontend: process.env["AQD_NODE_FRONTEND"] === "true",
  };
}
