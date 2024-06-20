import { ExternalLink } from "@tamagui/lucide-icons";
import { Anchor, H1, H2, H3, Paragraph, XStack, YStack, styled } from "tamagui";
const CodeH3 = styled(H3, { fontFamily: "$mono" });
const CenteredH1 = styled(H1, {
  fontWeight: "$2",
  textAlign: "center",
});
const CenteredH2 = styled(H2, {
  fontWeight: "$2",
  textAlign: "center",
  color: "$purple12",
});
const CenteredH3 = styled(H3, {
  fontWeight: "$2",
  textAlign: "center",
  color: "$purple12",
  fontSize: "$8",
});
export default function TabOneScreen() {
  return (
    <YStack f={1} ai="center" gap="$8" px="$10" pt="$5">
      <CenteredH1>Aquareum</CenteredH1>
      <CenteredH2>The Video Layer for Everything</CenteredH2>
      <YStack bg="$purple8" padding="$5" br="$3">
        <CenteredH3>World Premiere</CenteredH3>
        <CodeH3>Tuesday Jul 16 2024 10:00 PDT</CodeH3>
      </YStack>
    </YStack>
  );
}
