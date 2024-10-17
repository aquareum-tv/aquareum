// common types shared by players and controls and stuff
export type PlayerProps = {
  name: string;
  src: string;
  muted: boolean;
  fullscreen: boolean;
  protocol: string;
  showControls: boolean;
  telemetry: boolean;
  setMuted: (boolean) => void;
  setFullscreen: (boolean) => void;
  setProtocol: (string) => void;
  userInteraction: () => void;
  playerEvent: (
    e: Event,
    time: string,
    eventType: string,
    meta: { [key: string]: any },
  ) => void;
  playerId: string;
};

export type PlayerEvent = {
  id?: string;
  time: string;
  playerId: string;
  eventType: string;
  meta: { [key: string]: any };
};

export const PROTOCOL_HLS = "hls";
export const PROTOCOL_PROGRESSIVE_MP4 = "progressive-mp4";
export const PROTOCOL_PROGRESSIVE_WEBM = "progressive-webm";
