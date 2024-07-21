import { StatusBar } from "expo-status-bar";
import * as Updates from "expo-updates";
import { useEffect, useState } from "react";
import { Button, H2, H5, ScrollView, Text, View } from "tamagui";
import Constants from "expo-constants";
import { useToastController } from "@tamagui/toast";
import pkg from "../package.json";

export default function UpdatesDemo() {
  const version = pkg.version;
  const { currentlyRunning, isUpdateAvailable, isUpdatePending } =
    Updates.useUpdates();

  const [checked, setChecked] = useState(false);

  useEffect(() => {
    if (isUpdatePending) {
      Updates.reloadAsync();
    }
  }, [isUpdatePending]);

  useEffect(() => {
    if (isUpdateAvailable && checked) {
      Updates.fetchUpdateAsync();
    }
  }, [isUpdateAvailable, checked]);

  // If true, we show the button to download and run the update
  const showDownloadButton = isUpdateAvailable;
  const buttonText = isUpdateAvailable
    ? "Download new update"
    : "Check for updates";

  // Show whether or not we are running embedded code or an update
  let runTypeMessage = currentlyRunning.isEmbeddedLaunch ? "Bundled" : "OTA";
  if (currentlyRunning.isEmergencyLaunch) {
    runTypeMessage = "Recovery";
  }

  const toast = useToastController();

  return (
    <View f={1} alignItems="center" justifyContent="center" fg={1}>
      <View>
        <H2 textAlign="center">Aquareum v{version}</H2>
        <H5 textAlign="center" pb="$5">
          {runTypeMessage}
        </H5>
        <Button
          onPress={async () => {
            try {
              setChecked(true);
              const res = await Updates.checkForUpdateAsync();
              if (!res.isAvailable) {
                toast.show("No update found", {
                  message: "You are on the latest version of Aquareum, hooray!",
                });
              }
            } catch (e) {
              toast.show("Nope!", {
                message: e.message,
              });
            }
          }}
        >
          <Text>{buttonText}</Text>
        </Button>
      </View>
    </View>
  );
}
