import React, { useEffect, useRef } from "react";
import { useVideoPlayer, VideoView } from "expo-video";
import { View, Text, Button } from "tamagui";
import useAquareumNode from "hooks/useAquareumNode";

// export function Player() {
//   return <View f={1}></View>;
// }

export function Player(props: { src: string }) {
  const ref = useRef(null);
  const { url } = useAquareumNode();
  const src = `${url}/api/playback/${props.src}/stream.mp4`;
  const player = useVideoPlayer(src, (player) => {
    player.loop = true;
    player.play();
    player.muted = true;
  });

  useEffect(() => {
    const subscription = player.addListener("playingChange", (isPlaying) => {
      // setIsPlaying(isPlaying);
    });

    return () => {
      subscription.remove();
    };
  }, [player]);

  return (
    <View f={1} backgroundColor="green">
      <VideoView
        style={{ flex: 1, backgroundColor: "orange" }}
        ref={ref}
        player={player}
        allowsFullscreen
        allowsPictureInPicture
        nativeControls={false}
      />
      <View>
        {/* <Button
          title={isPlaying ? "Pause" : "Play"}
          onPress={() => {
            if (isPlaying) {
              player.pause();
            } else {
              player.play();
            }
            setIsPlaying(!isPlaying);
          }}
        /> */}
      </View>
    </View>
  );
}
