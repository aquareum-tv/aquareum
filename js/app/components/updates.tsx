import { StatusBar } from "expo-status-bar";
import * as Updates from "expo-updates";
import { useEffect } from "react";
import { Button, Text, View } from "react-native";

export default function UpdatesDemo() {
  const { currentlyRunning, isUpdateAvailable, isUpdatePending } =
    Updates.useUpdates();

  useEffect(() => {
    if (isUpdatePending) {
      // Update has successfully downloaded; apply it now
      Updates.reloadAsync();
    }
  }, [isUpdatePending]);

  // If true, we show the button to download and run the update
  const showDownloadButton = isUpdateAvailable;

  // Show whether or not we are running embedded code or an update
  const runTypeMessage = currentlyRunning.isEmbeddedLaunch
    ? "This app is running from built-in code"
    : "This app is running an update";

  return (
    <View>
      <Text>Updates Demo</Text>
      <Text>{runTypeMessage}</Text>
      <Button
        onPress={() => Updates.checkForUpdateAsync()}
        title="Check manually for updates"
      />
      {showDownloadButton ? (
        <Button
          onPress={() => Updates.fetchUpdateAsync()}
          title="Download and run update"
        />
      ) : null}
      <StatusBar style="auto" />
    </View>
  );
}
