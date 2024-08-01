/**aquareum
 * Welcome to Cloudflare Workers! This is your first worker.
 *
 * - Run `npm run dev` in your terminal to start a development server
 * - Open a browser tab at http://localhost:8787/ to see your worker in action
 * - Run `npm run deploy` to publish your worker
 *
 * Learn more at https://developers.cloudflare.com/workers/
 */

const re = new RegExp(
  `^aquareum-(v[0-9]\.[0-9]\.[0-9])(-[0-9a-f]+)?-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`,
);
const inputRe = new RegExp(`^aquareum-([0-9a-z]+)-([0-9a-z]+)\.(.+)$`);

const CI_API_V4_URL = "https://git.aquareum.tv/api/v4";
const DOWNLOAD_BASE_URL =
  "https://git.aquareum.tv/aquareum-tv/aquareum/-/package_files";
const PROJECT_ID = "1";

// Export a default object containing event handlers
export default {
  // The fetch handler is invoked when this worker receives a HTTP(S) request
  // and should return a Response (optionally wrapped in a Promise)
  async fetch(req, env, ctx) {
    const { pathname } = new URL(req.url);
    const [_, dl, branch, file] = pathname.split("/");
    if (!branch || !file) {
      return new Response("usage: /dl/latest/aquareum-linux-arm64.tar.gz", {
        status: 400,
      });
    }
    const inputPieces = inputRe.exec(file);
    {
      if (!inputPieces) {
        return new Response(`could not parse filename ${file}`);
      }
    }
    const [full, inputPlatform, inputArch, inputExt] = inputPieces;
    const packageUrl = `${CI_API_V4_URL}/projects/${PROJECT_ID}/packages?order_by=created_at&sort=desc&package_name=${branch}`;
    const packageReq = await fetch(packageUrl, {
      headers: {
        "user-agent": "aquareum-dl",
      },
    });
    const packages = await packageReq.json();
    const pkg = packages[0];
    if (!pkg) {
      return new Response(`package for branch ${branch} not found`, {
        status: 404,
      });
    }
    const fileUrl = `${CI_API_V4_URL}/projects/${PROJECT_ID}/packages/${pkg.id}/package_files`;
    console.log(fileUrl);
    const fileReq = await fetch(fileUrl);
    const files = await fileReq.json();
    let foundFile;
    let outUrl;
    for (const f of files) {
      const pieces = re.exec(f.file_name);
      if (!pieces) {
        console.error(`could not parse filename ${f.file_name}`);
        continue;
      }
      const [full, ver, hash, platform, arch, ext] = pieces;
      console.log({ full, ver, hash, platform, arch, ext });
      if (
        platform === inputPlatform &&
        arch === inputArch &&
        ext === inputExt
      ) {
        foundFile = f;
        const fullVer = `${ver}${hash ?? ""}`;
        outUrl = `${CI_API_V4_URL}/projects/${PROJECT_ID}/packages/generic/${branch}/${fullVer}/${f.file_name}`;
        break;
      }
    }
    if (!foundFile) {
      throw new Error(
        `could not find a file for platform=${inputPlatform} arch=${inputArch} ext=${inputExt})`,
      );
    }
    // const outUrl = `${DOWNLOAD_BASE_URL}/${foundFile.id}/download`;
    // "%s/projects/%s/packages/generic/%s/%s/aquareum-%s-linux-%s.tar.gz"
    // const outUrl
    return Response.redirect(outUrl, 302);
  },
};
