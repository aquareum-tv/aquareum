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
  variants: {
    small: {
      true: {
        fontSize: "$12",
        lineHeight: "$12",
      },
    },
  },
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
import { Countdown } from "components";
import { ImageBackground } from "react-native";
console.log(JSON.stringify(require(`assets/images/cube_small.png`)));
export default function TabOneScreen() {
  const { width, height } = useWindowDimensions();
  const small = width <= 600;
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
        <CenteredH1 padding="$5" small={small}>
          Aquareum
        </CenteredH1>
      </YStack>
      <View flexShrink={0} flexGrow={0} maxWidth="100%">
        <CenteredH2>The Video Layer for Everything</CenteredH2>
      </View>
      <Anchor href="https://docs.google.com/forms/d/e/1FAIpQLScA0O-qyrknM-p3jMNMHyA4Duld6TkusGUFTKwttnLmxWyhyQ/viewform?usp=sf_link">
        <View bg="rgb(189 110 134)" br="$3" padding="$5">
          <CenteredH3>Sign up for Updates</CenteredH3>
        </View>
      </Anchor>
      <View paddingBottom="$10">
        <Countdown to="2024-07-16T17:00:00.000Z" />
      </View>
    </YStack>
  );
}
