// import * as Popover from "@radix-ui/react-popover";
// import { ClipPayload } from "livepeer/dist/models/components";
// import { CheckIcon, ChevronDownIcon, XIcon } from "lucide-react";
import React, { useCallback, useEffect, useRef, useTransition } from "react";
import { Text, View } from "tamagui";
import WHEPClient from "./webrtc";
// import { createClip } from "./actions";
import { EXPO_PUBLIC_AQUAREUM_URL } from "constants/env";

export function Player(props: { src: string }) {
  const videoRef = useRef(null);
  useEffect(() => {
    if (!videoRef.current) {
      return;
    }
    const client = new WHEPClient(`/api/webrtc/${props.src}`, videoRef.current);
  }, [videoRef.current]);
  return (
    <View f={1} backgroundColor="#111">
      <video
        autoPlay={true}
        muted={true}
        ref={videoRef}
        loop={true}
        controls={true}
      />
    </View>
  );
}
