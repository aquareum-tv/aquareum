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
import { PlayerProps } from "./props";

export default function WebVideo(props: PlayerProps) {
  const [proto, setProto] = useState("hls");
  return <ProgressiveMP4Player src={props.src} muted={props.muted} />;
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

// export function HLSPlayer(props: { src: string }) {
//   const videoRef = useRef<HTMLVideoElement | null>(null);
//   const { url } = useAquareumNode();
//   useEffect(() => {
//     if (!videoRef.current) {
//       return;
//     }
//     const index = `${url}/api/hls/${props.src}/index.m3u8`;
//     if (Hls.isSupported()) {
//       var hls = new Hls();
//       hls.loadSource(index);
//       hls.attachMedia(videoRef.current);
//       hls.on(Hls.Events.MANIFEST_PARSED, () => {
//         if (!videoRef.current) {
//           return;
//         }
//         videoRef.current.play();
//       });
//       return () => {
//         hls.stopLoad();
//       };
//     } else if (videoRef.current.canPlayType("application/vnd.apple.mpegurl")) {
//       videoRef.current.src = index;
//       videoRef.current.addEventListener("canplay", () => {
//         if (!videoRef.current) {
//           return;
//         }
//         videoRef.current.play();
//       });
//     }
//   }, [videoRef.current]);
//   return <VideoElement ref={videoRef} />;
// }

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
