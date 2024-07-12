import { ExternalLink } from "@tamagui/lucide-icons";
import {
  Anchor,
  H1,
  H2,
  H3,
  Image,
  Paragraph,
  XStack,
  YStack,
  styled,
  View,
  Button,
  ScrollView,
  useWindowDimensions,
  isWeb,
  Text,
} from "tamagui";
import cube from "./cube.b64";
const CodeH3 = styled(H3, { fontFamily: "$mono" });
const CenteredH1 = styled(H1, {
  fontWeight: "$2",
  textAlign: "center",
  fontSize: isWeb ? "$16" : "$5",
  // flex: 1,
  lineHeight: "$16",
} as const);
const CenteredH2 = styled(H2, {
  fontWeight: "$2",
  textAlign: "center",
  // lineHeight: "$6",
  fontSize: "$10",
});
const CenteredH3 = styled(H3, {
  fontWeight: "$2",
  textAlign: "center",
  fontSize: "$8",
});
const CubeImage = styled(Image, {
  width: 100,
  height: 100,
  resizeMethod: "scale",
});
import { WebView } from "react-native-webview";
import { Countdown } from "components";
import { ImageBackground } from "react-native";
console.log(JSON.stringify(require(`assets/images/cube_small.png`)));
export default function TabOneScreen() {
  // const isLive = Date.now() >= 1721149200000;
  const isLive = false;
  return (
    <YStack f={1} ai="center" gap="$8" pt="$5" width="100%" alignItems="center">
      <YStack maxWidth="100%" width="100%" f={1} alignItems="center">
        <View fg={1} flexBasis={0} style={{ width: "100%" }}>
          <ImageBackground
            source={{ uri: cube }}
            style={{ width: "100%", height: "100%" }}
            resizeMode="contain"
          ></ImageBackground>
        </View>
      </YStack>
      <View flexShrink={0} flexGrow={0} maxWidth="100%">
        <CenteredH2>Aquareum: The Video Layer for Everything</CenteredH2>
      </View>
      <View fg={3} flexBasis={0} style={{ width: "100%" }}>
        <WebView
          scrollEnabled={false}
          source={{ uri: "https://iame.li" }}
          style={{ flex: 1, backgroundColor: "transparent" }}
        />
      </View>
      {!isLive && (
        <View>
          <Countdown to="2024-07-16T17:00:00.000Z" />
        </View>
      )}
      <View paddingBottom="$10"></View>
    </YStack>
  );
}
