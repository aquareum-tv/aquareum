// common types shared by players and controls and stuff
export type PlayerProps = {
  name: string;
  src: string;
  muted: boolean;
  setMuted: (boolean) => void;
  setFullscreen: (boolean) => void;
  fullscreen: boolean;
  protocol: string;
  setProtocol: (string) => void;
};

export const PROTOCOL_HLS = "hls";
export const PROTOCOL_PROGRESSIVE_MP4 = "progressive-mp4";
export const PROTOCOL_PROGRESSIVE_WEBM = "progressive-webm";
