import React, { useState } from "react";
import { Button, Text, View, XStack } from "tamagui";
import Controls from "./controls";
import Video from "./video";
import Fullscreen from "./fullscreen";
import { PlayerProps, PROTOCOL_HLS, PROTOCOL_PROGRESSIVE_MP4 } from "./props";
import usePlatform from "hooks/usePlatform";

export function Player(props: Partial<PlayerProps>) {
  if (typeof props.src !== "string") {
    return (
      <View>
        <Text>No source provided 🤷</Text>
      </View>
    );
  }
  const [muted, setMuted] = useState(true);
  const plat = usePlatform();
  let defProto = PROTOCOL_PROGRESSIVE_MP4;
  if (plat.isIOS) {
    defProto = PROTOCOL_HLS;
  } else if (plat.isSafari) {
    defProto = PROTOCOL_HLS;
  } else if (plat.isFirefox) {
    defProto = PROTOCOL_HLS;
  }
  const [protocol, setProtocol] = useState(defProto);
  const [fullscreen, setFullscreen] = useState(false);
  const childProps: PlayerProps = {
    name: props.name || props.src,
    src: props.src,
    muted: muted,
    setMuted: setMuted,
    setFullscreen: setFullscreen,
    fullscreen: fullscreen,
    protocol: protocol,
    setProtocol: setProtocol,
  };
  return (
    <View f={1} justifyContent="center" position="relative">
      <Fullscreen {...childProps}></Fullscreen>
    </View>
  );
}
