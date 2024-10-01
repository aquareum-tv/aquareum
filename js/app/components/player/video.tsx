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

export default function WebVideo(props: PlayerProps) {
  if (props.protocol === PROTOCOL_PROGRESSIVE_MP4) {
    return <ProgressiveMP4Player src={props.src} muted={props.muted} />;
  } else if (props.protocol === PROTOCOL_PROGRESSIVE_WEBM) {
    return <ProgressiveWebMPlayer src={props.src} muted={props.muted} />;
  } else if (props.protocol === PROTOCOL_HLS) {
    return <HLSPlayer src={props.src} muted={props.muted} />;
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

export function ProgressiveMP4Player(props: { src: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  return (
    <VideoElement
      muted={props.muted}
      ref={videoRef}
      src={`${url}/api/playback/${props.src}/stream.mp4`}
    />
  );
}

export function ProgressiveWebMPlayer(props: { src: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  return (
    <VideoElement
      muted={props.muted}
      ref={videoRef}
      src={`${url}/api/playback/${props.src}/stream.webm`}
    />
  );
}

export function HLSPlayer(props: { src: string; muted: boolean }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  useEffect(() => {
    if (!videoRef.current) {
      return;
    }
    const index = `${url}/api/playback/${props.src}/hls/stream.m3u8`;
    if (Hls.isSupported()) {
      var hls = new Hls();
      hls.loadSource(index);
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
      videoRef.current.src = index;
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
