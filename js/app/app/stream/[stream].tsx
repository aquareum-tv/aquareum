import { Player } from "components";
import { useLocalSearchParams } from "expo-router";
import { View } from "tamagui";

export default function StreamPage() {
  const params = useLocalSearchParams();
  if (typeof params.stream !== "string") {
    return <View />;
  }
  return <Player src={params.stream}></Player>;
}
