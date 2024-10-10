import {
  View,
  Text,
  XStack,
  Popover,
  YStack,
  Label,
  Input,
  Adapt,
  ListItem,
  Separator,
  YGroup,
} from "tamagui";
import {
  Volume2,
  VolumeX,
  Maximize,
  Minimize,
  Settings,
  ChevronRight,
  Moon,
  Star,
  Circle,
  CheckCircle,
  ChevronLeft,
  Sparkle,
} from "@tamagui/lucide-icons";
import { Animated, Button, Pressable, TouchableOpacity } from "react-native";
import { useEffect, useRef, useState } from "react";
import {
  PlayerProps,
  PROTOCOL_HLS,
  PROTOCOL_PROGRESSIVE_MP4,
  PROTOCOL_PROGRESSIVE_WEBM,
} from "./props";

const Bar = (props) => (
  <XStack
    height={50}
    backgroundColor="rgba(0,0,0,0.8)"
    justifyContent="space-between"
    flex-direction="row"
  >
    {props.children}
  </XStack>
);

const Part = (props) => (
  <View alignItems="stretch" justifyContent="center" flexDirection="row">
    {props.children}
  </View>
);

export default function Controls(props: PlayerProps) {
  const fadeAnim = useRef(new Animated.Value(1)).current;

  // useEffect(() => {
  //   Animated.timing(fadeAnim, {
  //     toValue: props.showControls ? 1 : 1,
  //     duration: 175,
  //     useNativeDriver: false,
  //   }).start();
  // }, [fadeAnim, props.showControls]);

  let cursor = {};
  if (props.fullscreen && !props.showControls) {
    cursor = { cursor: "none" };
  }

  return (
    <View
      position="absolute"
      width="100%"
      height="100%"
      zIndex={999}
      flexDirection="column"
      justifyContent="space-between"
      animation="quick"
      animateOnly={["opacity"]}
      opacity={props.showControls ? 1 : 0}
      onPointerMove={props.userInteraction}
      onTouchStart={props.userInteraction}
      {...cursor}
    >
      {/* <Animated.View
        // onPointerMove={props.userInteraction}
        // onTouchStart={props.userInteraction}
        style={{
          flex: 1,
          opacity: fadeAnim,
          width: "100%",
          height: "100%",
          flexDirection: "column",
          justifyContent: "space-between",
        }}
      > */}
      <Bar>
        <Part>
          <View justifyContent="center" paddingLeft="$5">
            <Text>{props.name}</Text>
          </View>
        </Part>
        <Part>{/* <Text>Top Right</Text> */}</Part>
      </Bar>
      <Bar>
        <Part>
          <Pressable
            style={{
              justifyContent: "center",
            }}
            onPress={() => props.setMuted(!props.muted)}
          >
            <View paddingLeft="$5" paddingRight="$3" justifyContent="center">
              {props.muted ? <VolumeX></VolumeX> : <Volume2></Volume2>}
            </View>
          </Pressable>
        </Part>
        <Part>
          <PopoverMenu {...props} />
          <Pressable
            style={{
              justifyContent: "center",
            }}
            onPress={() => props.setFullscreen(!props.fullscreen)}
          >
            <View paddingLeft="$3" paddingRight="$5" justifyContent="center">
              {props.fullscreen ? <Minimize /> : <Maximize />}
            </View>
          </Pressable>
        </Part>
      </Bar>
      {/* </Animated.View> */}
    </View>
  );
}

export function PopoverMenu(props: PlayerProps) {
  return (
    <Popover
      size="$5"
      allowFlip
      placement="top"
      keepChildrenMounted
      stayInFrame
    >
      <Popover.Trigger asChild cursor="pointer">
        <View paddingLeft="$3" paddingRight="$5" justifyContent="center">
          <Settings />
        </View>
      </Popover.Trigger>

      <Adapt when="sm" platform="touch">
        <Popover.Sheet modal dismissOnSnapToBottom snapPoints={[50]}>
          <Popover.Sheet.Frame padding="$2">
            <Adapt.Contents />
          </Popover.Sheet.Frame>
          <Popover.Sheet.Overlay
            animation="lazy"
            enterStyle={{ opacity: 0 }}
            exitStyle={{ opacity: 0 }}
          />
        </Popover.Sheet>
      </Adapt>

      <Popover.Content
        borderWidth={0}
        padding="$0"
        enterStyle={{ y: -10, opacity: 0 }}
        exitStyle={{ y: -10, opacity: 0 }}
        elevate
        userSelect="none"
        animation={[
          "quick",
          {
            opacity: {
              overshootClamping: true,
            },
          },
        ]}
      >
        <GearMenu {...props} />
      </Popover.Content>
    </Popover>
  );
}

function GearMenu(props: PlayerProps) {
  const [menu, setMenu] = useState("root");
  return (
    <YGroup alignSelf="center" bordered width={240} size="$5" borderRadius="$0">
      {menu == "root" && (
        <>
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="Playback Protocol"
              subTitle="How play?"
              icon={Star}
              iconAfter={ChevronRight}
              onPress={() => setMenu("protocol")}
            />
          </YGroup.Item>
          <Separator />
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="Quality"
              subTitle="WIP"
              icon={Sparkle}
              iconAfter={ChevronRight}
            />
          </YGroup.Item>
        </>
      )}
      {menu == "protocol" && (
        <>
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="Back"
              icon={ChevronLeft}
              onPress={() => setMenu("root")}
            />
          </YGroup.Item>
          <Separator />
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="HLS"
              subTitle="HTTP Live Streaming"
              icon={Star}
              iconAfter={props.protocol === PROTOCOL_HLS ? CheckCircle : Circle}
              onPress={() => props.setProtocol(PROTOCOL_HLS)}
            />
          </YGroup.Item>
          <Separator />
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="Progressive MP4"
              subTitle="MP4 but loooong"
              icon={Moon}
              iconAfter={
                props.protocol === PROTOCOL_PROGRESSIVE_MP4
                  ? CheckCircle
                  : Circle
              }
              onPress={() => props.setProtocol(PROTOCOL_PROGRESSIVE_MP4)}
            />
          </YGroup.Item>
          <Separator />
          <YGroup.Item>
            <ListItem
              hoverTheme
              pressTheme
              title="Progressive WebM"
              subTitle="WebM but loooong"
              icon={Moon}
              iconAfter={
                props.protocol === PROTOCOL_PROGRESSIVE_WEBM
                  ? CheckCircle
                  : Circle
              }
              onPress={() => props.setProtocol(PROTOCOL_PROGRESSIVE_WEBM)}
            />
          </YGroup.Item>
        </>
      )}
    </YGroup>
  );
}
