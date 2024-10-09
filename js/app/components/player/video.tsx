import React, {
  ForwardedRef,
  forwardRef,
  useCallback,
  useEffect,
  useRef,
  useState,
  useTransition,
} from "react";
import { Button, Text, View, XStack } from "tamagui";
import WHEPClient from "./webrtc";
import Hls from "hls.js";
import { Circle, CheckCircle } from "@tamagui/lucide-icons";
import useAquareumNode from "hooks/useAquareumNode";
import Controls from "./controls";
import {
  PlayerProps,
  PROTOCOL_HLS,
  PROTOCOL_PROGRESSIVE_MP4,
  PROTOCOL_PROGRESSIVE_WEBM,
} from "./props";
import { srcToUrl } from "./shared";

export default function WebVideo(props: PlayerProps) {
  const { url, protocol } = srcToUrl(props);
  console.log("got", url, protocol);
  if (protocol === PROTOCOL_PROGRESSIVE_MP4) {
    return <ProgressiveMP4Player url={url} muted={props.muted} />;
  } else if (protocol === PROTOCOL_PROGRESSIVE_WEBM) {
    return <ProgressiveWebMPlayer url={url} muted={props.muted} />;
  } else if (protocol === PROTOCOL_HLS) {
    return <HLSPlayer url={url} muted={props.muted} />;
  } else {
    throw new Error(`unknown playback protocol ${props.protocol}`);
  }
}

const VideoElement = forwardRef(
  (
    props: { src?: string; muted: boolean },
    ref: ForwardedRef<HTMLVideoElement>,
  ) => {
    return (
      <View backgroundColor="#111" alignItems="stretch" f={1}>
        <video
          autoPlay={true}
          ref={ref}
          loop={true}
          controls={false}
          src={props.src}
          muted={props.muted}
          crossOrigin="anonymous"
          style={{
            objectFit: "contain",
            backgroundColor: "transparent",
            width: "100%",
            height: "100%",
          }}
        />
      </View>
    );
  },
);

export function ProgressiveMP4Player(props: { url: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  return (
    <VideoElement
      muted={props.muted}
      ref={videoRef}
      src={`${url}/api/playback/${props.url}/stream.mp4`}
    />
  );
}

export function ProgressiveWebMPlayer(props: { url: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  return (
    <VideoElement
      muted={props.muted}
      ref={videoRef}
      src={`${url}/api/playback/${props.url}/stream.webm`}
    />
  );
}

export function HLSPlayer(props: { url: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  useEffect(() => {
    if (!videoRef.current) {
      return;
    }
    if (Hls.isSupported()) {
      var hls = new Hls();
      hls.loadSource(props.url);
      hls.attachMedia(videoRef.current);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (!videoRef.current) {
          return;
        }
        videoRef.current.play();
      });
      return () => {
        hls.stopLoad();
      };
    } else if (videoRef.current.canPlayType("application/vnd.apple.mpegurl")) {
      videoRef.current.src = props.url;
      videoRef.current.addEventListener("canplay", () => {
        if (!videoRef.current) {
          return;
        }
        videoRef.current.play();
      });
    }
  }, [videoRef.current]);
  return <VideoElement ref={videoRef} muted={props.muted} />;
}

// export function WebRTCPlayer(props: { src: string }) {
//   const videoRef = useRef<HTMLVideoElement | null>(null);
//   const { url } = useAquareumNode();
//   useEffect(() => {
//     if (!videoRef.current) {
//       return;
//     }
//     const client = new WHEPClient(
//       `${url}/api/webrtc/${props.src}`,
//       videoRef.current,
//     );
//     return () => {
//       client.close();
//     };
//   }, [videoRef.current]);
//   return <VideoElement ref={videoRef} />;
// }
