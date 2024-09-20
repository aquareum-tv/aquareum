import os from "os";
import { resolve } from "path";
import { access, constants } from "fs/promises";
import { spawn } from "child_process";
import getEnv from "./env";

const findExe = async (): Promise<string> => {
  const { isDev } = getEnv();
  let fname = "aquareum";
  if (os.platform() === "win32") {
    fname += ".exe";
  }
  let exe: string;
  if (isDev) {
    // theoretically cwd is aquareum/js/desktop:
    exe = resolve(process.cwd(), "..", "..", "bin", fname);
  } else {
    exe = resolve(process.resourcesPath, fname);
  }
  try {
    await access(exe, constants.F_OK);
  } catch (e) {
    throw new Error(
      `could not find aquareum node binary at ${exe}: ${e.message}`,
    );
  }
  return exe;
};

export default async function makeNode() {
  const exe = await findExe();
  const addr = "127.0.0.1:38082";
  const internalAddr = "127.0.0.1:39092";
  const proc = spawn(exe, ["--insecure"], {
    stdio: "inherit",
    env: {
      ...process.env,
      AQ_NO_MIST: "true",
      AQ_HTTP_ADDR: addr,
      AQ_HTTP_INTERNAL_ADDR: internalAddr,
    },
  });
  await checkService(`http://${addr}/api/healthz`);

  return {
    proc,
    addr: `http://${addr}`,
  };
}

const checkService = (
  url: string,
  interval = 300,
  timeout = 10000,
): Promise<void> => {
  let attempts = 0;
  const maxAttempts = timeout / interval;

  return new Promise((resolve, reject) => {
    const intervalId = setInterval(async () => {
      attempts++;

      try {
        const response = await fetch(url);
        if (response.ok) {
          // Response status in the range 200-299
          clearInterval(intervalId);
          resolve();
        }
      } catch (error) {
        // Fetch failed, continue trying
      }

      if (attempts >= maxAttempts) {
        clearInterval(intervalId);
        reject(new Error("aquareum did not boot up in time"));
      }
    }, interval);
  });
};
