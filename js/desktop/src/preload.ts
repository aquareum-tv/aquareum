import { contextBridge } from "electron";

contextBridge.exposeInMainWorld("AQ_ELECTRON", {
  isElectron: true,
});
