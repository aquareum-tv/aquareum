// common types shared by players and controls and stuff
export type PlayerProps = {
  name: string;
  src: string;
  muted: boolean;
  setMuted: (boolean) => void;
  setFullscreen: (boolean) => void;
  fullscreen: boolean;
};
