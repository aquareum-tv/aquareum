import UpdatesDemo from "components/updates";
import { isWeb, Text, View } from "tamagui";

export default function OptionsScreen() {
  if (isWeb) {
    return <View />;
  }
  return (
    <View flex={1}>
      <UpdatesDemo />
    </View>
  );
}
