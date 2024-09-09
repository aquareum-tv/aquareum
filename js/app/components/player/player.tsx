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

export function Player(props: { src: string }) {
  const [proto, setProto] = useState("hls");
  let p;
  if (proto === "webrtc") {
    p = <WebRTCPlayer src={props.src} />;
  } else if (proto === "hls") {
    p = <HLSPlayer src={props.src} />;
  }
  return (
    <View f={1}>
      {p}
      <XStack justifyContent="center">
        <PickerButton
          name="webrtc"
          title="WebRTC"
          picked={proto}
          setProto={setProto}
        />
        <PickerButton
          name="hls"
          title="HLS"
          picked={proto}
          setProto={setProto}
        />
      </XStack>
    </View>
  );
}

const PickerButton = (props: {
  name: string;
  picked: string;
  title: string;
  setProto: (string) => void;
}) => {
  const on = props.picked === props.name;
  return (
    <Button
      disabled={on}
      margin="$3"
      opacity={on ? 0.5 : 1}
      icon={on ? CheckCircle : Circle}
      onPress={() => props.setProto(props.name)}
    >
      {props.title}
    </Button>
  );
};

export function HLSPlayer(props: { src: string }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  useEffect(() => {
    if (!videoRef.current) {
      return;
    }
    const index = `${url}/api/hls/${props.src}/index.m3u8`;
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
  return <VideoElement ref={videoRef} />;
}

export function WebRTCPlayer(props: { src: string }) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const { url } = useAquareumNode();
  useEffect(() => {
    if (!videoRef.current) {
      return;
    }
    const client = new WHEPClient(
      `${url}/api/webrtc/${props.src}`,
      videoRef.current,
    );
    return () => {
      client.close();
    };
  }, [videoRef.current]);
  return <VideoElement ref={videoRef} />;
}

const VideoElement = forwardRef(
  (props, ref: ForwardedRef<HTMLVideoElement>) => {
    return (
      <View backgroundColor="#111">
        <video autoPlay={true} ref={ref} loop={true} controls={true} />
      </View>
    );
  },
);
