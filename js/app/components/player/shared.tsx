import useAquareumNode from "hooks/useAquareumNode";
import {
  PlayerProps,
  PROTOCOL_HLS,
  PROTOCOL_PROGRESSIVE_MP4,
  PROTOCOL_PROGRESSIVE_WEBM,
} from "./props";

const protocolSuffixes = {
  m3u8: PROTOCOL_HLS,
  mp4: PROTOCOL_PROGRESSIVE_MP4,
  webm: PROTOCOL_PROGRESSIVE_WEBM,
};

export function srcToUrl(props: PlayerProps): {
  url: string;
  protocol: string;
} {
  if (props.src.startsWith("http://") || props.src.startsWith("https://")) {
    const suffix = props.src.split(".").pop() as string;
    if (protocolSuffixes[suffix]) {
      console.log(`found ${protocolSuffixes[suffix]}`);
      return {
        url: props.src,
        protocol: protocolSuffixes[suffix],
      };
    } else {
      throw new Error(`unknown playback protocol: ${suffix}`);
    }
  }
  const { url } = useAquareumNode();
  let outUrl;
  if (props.protocol === PROTOCOL_HLS) {
    outUrl = `${url}/api/playback/${props.src}/hls/stream.m3u8`;
  } else if (props.protocol === PROTOCOL_PROGRESSIVE_MP4) {
    outUrl = `${url}/api/playback/${props.src}/stream.mp4`;
  } else if (props.protocol === PROTOCOL_PROGRESSIVE_WEBM) {
    outUrl = `${url}/api/playback/${props.src}/stream.webm`;
  } else {
    throw new Error(`unknown playback protocol: ${url}`);
  }
  return {
    protocol: props.protocol,
    url: outUrl,
  };
}
