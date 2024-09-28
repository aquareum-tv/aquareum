import React, { useEffect, useRef } from "react";
import { useVideoPlayer, VideoView } from "expo-video";
import { View, Text, Button } from "tamagui";
import useAquareumNode from "hooks/useAquareumNode";
import Controls from "./controls";
import { Platform } from "react-native";
import { PlayerProps } from "./props";

// export function Player() {
//   return <View f={1}></View>;
// }

export default function NativeVideo(
  props: PlayerProps & { videoRef: React.RefObject<VideoView> },
) {
  const { url } = useAquareumNode();
  const src = `${url}/api/playback/${props.src}/stream.mp4`;
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

  // useEffect(() => {
  //   if (!ref.current) {
  //     return;
  //   }
  //   if (props.fullscreen) {
  //     ref.current.enterFullscreen();
  //   }
  //   if (!props.fullscreen) {
  //     if (Platform.OS !== "android") {
  //       ref.current.exitFullscreen();
  //     }
  //   }
  // }, [props.fullscreen, ref.current]);

  return (
    <VideoView
      style={{ flex: 1, backgroundColor: "#111" }}
      ref={props.videoRef}
      player={player}
      allowsFullscreen
      allowsPictureInPicture
      nativeControls={false}
      onFullscreenEnter={() => {
        props.setFullscreen(true);
      }}
      onFullscreenExit={() => {
        props.setFullscreen(false);
      }}
    />
  );
}
