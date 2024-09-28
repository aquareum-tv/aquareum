import { View, Text, XStack } from "tamagui";
import { Volume2, VolumeX, Maximize, Minimize } from "@tamagui/lucide-icons";
import { Pressable, TouchableOpacity } from "react-native";

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

export type Controls = {
  name: string;
  muted: boolean;
  setMuted: (boolean) => void;
  setFullscreen: (boolean) => void;
  fullscreen: boolean;
};

export default function Controls(props: Controls) {
  return (
    <View
      position="absolute"
      width="100%"
      height="100%"
      zIndex={999}
      flexDirection="column"
      justifyContent="space-between"
    >
      <Bar>
        <Part>
          <View justifyContent="center" paddingLeft="$5">
            <Text>{props.name}</Text>
          </View>
        </Part>
        <Part>
          <Text>Top Right</Text>
        </Part>
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
    </View>
  );
}
