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
} from "tamagui";
const CodeH3 = styled(H3, { fontFamily: "$mono" });
const CenteredH1 = styled(H1, {
  fontWeight: "$2",
  textAlign: "center",
  fontSize: "$16",
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
  lineHeight: "$6",
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
export default function TabOneScreen() {
  const { width, height } = useWindowDimensions();
  const small = width <= 600;
  return (
    <ScrollView flex={1}>
      <YStack f={1} ai="center" gap="$8" px="$10" pt="$5" maxWidth="100%">
        <CubeImage
          source={
            small
              ? require(`../../assets/images/cube_small.png`)
              : require(`../../assets/images/cube.png`)
          }
          resizeMethod="scale"
          scale={0.5}
        />
        <CenteredH1 padding="$5" small={small}>
          Aquareum
        </CenteredH1>
        <CenteredH2>The Video Layer for Everything</CenteredH2>
        <XStack alignItems="center" justifyContent="space-around"></XStack>
        <Anchor href="https://docs.google.com/forms/d/e/1FAIpQLScA0O-qyrknM-p3jMNMHyA4Duld6TkusGUFTKwttnLmxWyhyQ/viewform?usp=sf_link">
          <View mr="$4" bg="rgb(189 110 134)" br="$3" padding="$5">
            <CenteredH3>Sign up for Updates</CenteredH3>
          </View>
        </Anchor>
        <Countdown to="2024-07-16T17:00:00.000Z" />
      </YStack>
    </ScrollView>
  );
}
