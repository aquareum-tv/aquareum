import type { ForgeConfig } from "@electron-forge/shared-types";
import { MakerSquirrel } from "@electron-forge/maker-squirrel";
import { MakerDMG } from "@electron-forge/maker-dmg";
import { MakerDeb } from "@electron-forge/maker-deb";
import { MakerRpm } from "@electron-forge/maker-rpm";
import { AutoUnpackNativesPlugin } from "@electron-forge/plugin-auto-unpack-natives";
import { WebpackPlugin } from "@electron-forge/plugin-webpack";
import { FusesPlugin } from "@electron-forge/plugin-fuses";
import { FuseV1Options, FuseVersion } from "@electron/fuses";

import { mainConfig } from "./webpack.main.config";
import { rendererConfig } from "./webpack.renderer.config";
import fs from "fs";

const config: ForgeConfig = {
  packagerConfig: {
    asar: true,
    name: "Aquareum",
    appVersion: process.env.VERSION,
    buildVersion: process.env.VERSION,
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
      packageJson.version = process.env.VERSION;
      return packageJson;
    },
  },
  rebuildConfig: {},
  makers: [
    new MakerSquirrel({}),
    new MakerDMG({}, ["darwin"]),
    // new MakerRpm({}),
    new MakerDeb({}),
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
};

export default config;
