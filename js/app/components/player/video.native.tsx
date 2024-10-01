import React, { useEffect } from "react";
import { useVideoPlayer, VideoView } from "expo-video";
import useAquareumNode from "hooks/useAquareumNode";
import {
  PlayerProps,
  PROTOCOL_HLS,
  PROTOCOL_PROGRESSIVE_MP4,
  PROTOCOL_PROGRESSIVE_WEBM,
} from "./props";

// export function Player() {
//   return <View f={1}></View>;
// }

export default function NativeVideo(
  props: PlayerProps & { videoRef: React.RefObject<VideoView> },
) {
  const { url } = useAquareumNode();
  let src: string;
  if (props.protocol === PROTOCOL_HLS) {
    src = `${url}/api/playback/${props.src}/hls/stream.m3u8`;
  } else if (props.protocol === PROTOCOL_PROGRESSIVE_MP4) {
    src = `${url}/api/playback/${props.src}/stream.mp4`;
  } else if (props.protocol === PROTOCOL_PROGRESSIVE_WEBM) {
    src = `${url}/api/playback/${props.src}/stream.webm`;
  } else {
    throw new Error(`unknown playback protocol: ${url}`);
  }
  const player = useVideoPlayer(src, (player) => {
    player.loop = true;
    player.muted = props.muted;
    player.play();
  });

  useEffect(() => {
    player.muted = props.muted;
  }, [props.muted, player]);

  useEffect(() => {
    const subscription = player.addListener("playingChange", (isPlaying) => {
      // setIsPlaying(isPlaying);
    });

    return () => {
      subscription.remove();
    };
  }, [player]);

  return (
    <VideoView
      style={{ flex: 1, backgroundColor: "#111" }}
      ref={props.videoRef}
      player={player}
      allowsFullscreen
      allowsPictureInPicture
      nativeControls={props.fullscreen}
      onFullscreenEnter={() => {
        props.setFullscreen(true);
      }}
      onFullscreenExit={() => {
        props.setFullscreen(false);
      }}
    />
  );
}
