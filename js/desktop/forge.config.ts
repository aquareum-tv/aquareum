import type { ForgeConfig } from "@electron-forge/shared-types";
import { MakerSquirrel } from "@electron-forge/maker-squirrel";
import { MakerDMG } from "@electron-forge/maker-dmg";
import { MakerZIP } from "@electron-forge/maker-zip";
// import { MakerDeb } from "@electron-forge/maker-deb";
// import { MakerRpm } from "@electron-forge/maker-rpm";
import { AutoUnpackNativesPlugin } from "@electron-forge/plugin-auto-unpack-natives";
import { WebpackPlugin } from "@electron-forge/plugin-webpack";
import { FusesPlugin } from "@electron-forge/plugin-fuses";
import { FuseV1Options, FuseVersion } from "@electron/fuses";
import { PublisherS3 } from "@electron-forge/publisher-s3";
import { mainConfig } from "./webpack.main.config";
import { rendererConfig } from "./webpack.renderer.config";
import fs from "fs";
import child_process from "child_process";
import { MakerAppImage } from "@reforged/maker-appimage";

export default async function () {
  // go get the version from the actual script
  let version = child_process
    .execSync("go run ../../pkg/config/git/git.go -v")
    .toString()
    .slice(1);

  // https://github.com/Squirrel/Squirrel.Windows/issues/1394#issuecomment-2356692821
  // This makes the Desktop app have confusing numbers, but actual relases just use X.Y.Z
  // so it doesn't really matter. Those are just for testing.
  if (version.includes("-")) {
    version = version.replace("-", "-z");
  }

  const config: ForgeConfig = {
    packagerConfig: {
      asar: true,
      name: "Aquareum",
      appVersion: version,
      buildVersion: version,
    },
    hooks: {
      prePackage: async (config, plat, arch) => {
        let platform = plat;
        let architecture = arch;
        if (platform === "win32") {
          platform = "windows";
        }
        if (architecture === "x64") {
          architecture = "amd64";
        }
        let binary = `../../build-${platform}-${architecture}/aquareum`;
        if (platform === "windows") {
          binary += ".exe";
        }
        const exists = fs.existsSync(binary);
        if (!exists) {
          throw new Error(
            `could not find ${binary} while building electron bundle. do you need to run make ${platform}-${architecture}?`,
          );
        }

        if (!config.packagerConfig) {
          throw new Error("config.packageConfig undefined");
        }
        config.packagerConfig.extraResource = [binary];
      },
      readPackageJson: async (forgeConfig, packageJson) => {
        packageJson.version = version;
        return packageJson;
      },
    },
    rebuildConfig: {},
    makers: [
      new MakerSquirrel({}),
      new MakerDMG({}, ["darwin"]),
      new MakerZIP(
        (arch) => {
          return {
            // Note that we must provide this S3 URL here
            // in order to support smooth version transitions
            // especially when using a CDN to front your updates
            macUpdateManifestBaseUrl: `https://1097-169-197-143-250.ngrok-free.app/aquareum/aquareum-desktop/darwin/${arch}`,
          };
        },
        ["darwin"],
      ),
      // new MakerRpm({}),
      new MakerAppImage({
        options: {
          bin: "Aquareum",
        },
      }),
    ],
    plugins: [
      new AutoUnpackNativesPlugin({}),
      new WebpackPlugin({
        mainConfig,
        renderer: {
          config: rendererConfig,
          entryPoints: [
            {
              html: "./src/index.html",
              js: "./src/renderer.ts",
              name: "main_window",
              preload: {
                js: "./src/preload.ts",
              },
            },
          ],
        },
      }),
      // Fuses are used to enable/disable various Electron functionality
      // at package time, before code signing the application
      new FusesPlugin({
        version: FuseVersion.V1,
        [FuseV1Options.RunAsNode]: false,
        [FuseV1Options.EnableCookieEncryption]: true,
        [FuseV1Options.EnableNodeOptionsEnvironmentVariable]: false,
        [FuseV1Options.EnableNodeCliInspectArguments]: false,
        [FuseV1Options.EnableEmbeddedAsarIntegrityValidation]: true,
        [FuseV1Options.OnlyLoadAppFromAsar]: true,
      }),
    ],
    publishers: [
      new PublisherS3({
        endpoint: "http://192.168.8.136:9000",
        accessKeyId: "minioadmin",
        secretAccessKey: "minioadmin",
        public: true,
        bucket: "aquareum",
        region: "unused",
      }),
    ],
  };
  return config;
}
