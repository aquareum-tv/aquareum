import React, { useEffect } from "react";
import { useVideoPlayer, VideoView } from "expo-video";
import useAquareumNode from "hooks/useAquareumNode";
import {
  PlayerProps,
  PROTOCOL_HLS,
  PROTOCOL_PROGRESSIVE_MP4,
  PROTOCOL_PROGRESSIVE_WEBM,
} from "./props";
import { srcToUrl } from "./shared";

// export function Player() {
//   return <View f={1}></View>;
// }

export default function NativeVideo(
  props: PlayerProps & { videoRef: React.RefObject<VideoView> },
) {
  const { url } = srcToUrl(props);
  const player = useVideoPlayer(url, (player) => {
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
