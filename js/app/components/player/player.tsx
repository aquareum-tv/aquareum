import React, { useState } from "react";
import { Button, Text, View, XStack } from "tamagui";
import Controls from "./controls";
import Video from "./video";
import Fullscreen from "./fullscreen";
import { PlayerProps } from "./props";

export function Player(props: { src: string }) {
  const [muted, setMuted] = useState(true);
  const [fullscreen, setFullscreen] = useState(false);
  const childProps: PlayerProps = {
    name: props.src,
    src: props.src,
    muted: muted,
    setMuted: setMuted,
    setFullscreen: setFullscreen,
    fullscreen: fullscreen,
  };
  return (
    <View f={1} justifyContent="center" position="relative">
      <Fullscreen {...childProps}></Fullscreen>
    </View>
  );
}
