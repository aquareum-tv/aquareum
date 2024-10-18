import "source-map-support/register";
import { app, BrowserWindow, autoUpdater, dialog } from "electron";
// This allows TypeScript to pick up the magic constants that's auto-generated by Forge's Webpack
// plugin that tells the Electron app where to look for the Webpack-bundled app code (depending on
// whether you're running in development or production).
declare const MAIN_WINDOW_WEBPACK_ENTRY: string;
declare const MAIN_WINDOW_PRELOAD_WEBPACK_ENTRY: string;
import makeNode from "./node";
import getEnv from "./env";
import initUpdater from "./updater";
import path, { resolve } from "path";
import { parseArgs } from "node:util";
import { v7 as uuidv7 } from "uuid";

// Handle creating/removing shortcuts on Windows when installing/uninstalling.
if (require("electron-squirrel-startup")) {
  app.quit();
} else {
  const { values: args, positionals } = parseArgs({
    options: {
      path: {
        type: "string",
        default: "",
      },
      "self-test": {
        type: "boolean",
      },
      "self-test-duration": {
        type: "string",
        default: "60000",
      },
    },
  });
  const env = getEnv();
  console.log(
    "starting with: ",
    JSON.stringify({ args, positionals, env }, null, 2),
  );

  app.on("ready", async () => {
    try {
      if (!args["self-test"]) {
        await start();
      } else {
        await runSelfTest();
      }
    } catch (e) {
      console.error(e);
      const dialogOpts: Electron.MessageBoxOptions = {
        type: "info",
        buttons: ["Quit Aquareum"],
        title: "Error on Bootup",
        message: "Please report to the Aquareum developers at git.aquareum.tv!",
        detail: e.message + "\n" + e.stack,
      };

      await dialog.showMessageBox(dialogOpts);
      app.quit();
    }
  });

  const makeWindow = async (): Promise<BrowserWindow> => {
    const { isDev } = getEnv();
    let logoFile: string;
    if (isDev) {
      // theoretically cwd is aquareum/js/desktop:
      logoFile = resolve(process.cwd(), "assets", "aquareum-logo.png");
    } else {
      logoFile = resolve(process.resourcesPath, "aquareum-logo.png");
    }
    const window = new BrowserWindow({
      height: 600,
      width: 800,
      icon: logoFile,
      webPreferences: {
        preload: MAIN_WINDOW_PRELOAD_WEBPACK_ENTRY,
      },
      // titleBarStyle: "hidden",
      // titleBarOverlay: true,
    });

    window.removeMenu();

    return window;
  };

  const start = async (): Promise<void> => {
    initUpdater();
    const { skipNode, nodeFrontend } = getEnv();
    let loadAddr;
    if (!skipNode) {
      const { addr } = await makeNode({ env: {}, autoQuit: true });
      loadAddr = addr;
    }
    const mainWindow = await makeWindow();

    let startPath;
    if (nodeFrontend) {
      startPath = `${loadAddr}${args.path}`;
    } else {
      startPath = `http://localhost:38081${args.path}`;
    }
    console.log(`opening ${startPath}`);
    mainWindow.loadURL(startPath);
  };

  const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));
  // how much of our time is spent playing for a success?
  const PLAYING_SUCCESS = 0.9;

  const runSelfTest = async (): Promise<void> => {
    let exitCode = 0;
    const { addr, internalAddr, proc } = await makeNode({
      env: {
        AQ_TEST_STREAM: "true",
      },
      autoQuit: false,
    });
    try {
      const mainWindow = await makeWindow();

      const testId = uuidv7();
      const definitions = [
        {
          name: "hls",
          src: "/hls/stream.m3u8",
        },
        {
          name: "progressive-mp4",
          src: "/stream.mp4",
        },
        {
          name: "progressive-webm",
          src: "/stream.webm",
        },
      ];
      const tests = definitions.map((x) => ({
        name: x.name,
        playerId: `${testId}-${x.name}`,
        src: `${addr}/api/playback/self-test${x.src}`,
        showControls: true,
      }));
      const enc = encodeURIComponent(JSON.stringify(tests));

      mainWindow.loadURL(`${addr}/multi/${enc}`);

      await delay(parseInt(args["self-test-duration"]));
      const reports = await Promise.all(
        tests.map(async (t) => {
          const res = await fetch(
            `${internalAddr}/player-report/${t.playerId}`,
          );
          const data = (await res.json()) as { [k: string]: number };
          return { ...t, data: data };
        }),
      );
      let failed = false;
      const percentages = reports.map((report) => {
        let total = 0;
        for (const [state, ms] of Object.entries(report.data)) {
          total += ms;
        }
        const pcts: { [k: string]: number } = { playing: 0 };
        for (const [state, ms] of Object.entries(report.data)) {
          pcts[state] = ms / total;
        }
        if (pcts.playing < PLAYING_SUCCESS) {
          failed = true;
        }
        return { ...report, pcts };
      });
      console.log(JSON.stringify(percentages, null, 2));
      if (failed) {
        console.log("test failed! exiting 1");
        exitCode = 1;
      }
    } catch (e) {
      console.error(`error in self-test: ${e}`);
    } finally {
      proc.kill("SIGTERM");
      app.exit(exitCode);
    }
  };

  // This method will be called when Electron has finished
  // initialization and is ready to create browser windows.
  // Some APIs can only be used after this event occurs.

  // Quit when all windows are closed, except on macOS. There, it's common
  // for applications and their menu bar to stay active until the user quits
  // explicitly with Cmd + Q.
  // app.on("window-all-closed", () => {
  //   if (process.platform !== "darwin") {
  //     app.quit();
  //   }
  // });

  app.on("activate", () => {
    // On OS X it's common to re-create a window in the app when the
    // dock icon is clicked and there are no other windows open.
    // if (BrowserWindow.getAllWindows().length === 0) {
    // }
  });

  // In this file you can include the rest of your app's specific main process
  // code. You can also put them in separate files and import them here.
}
