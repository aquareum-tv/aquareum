import { StatusBar } from "expo-status-bar";
import * as Updates from "expo-updates";
import { useEffect, useState } from "react";
import { Button, ScrollView, Text, View } from "tamagui";

export default function UpdatesDemo() {
  const { currentlyRunning, isUpdateAvailable, isUpdatePending } =
    Updates.useUpdates();

  const [debugLogs, setDebugLogs] = useState<string[]>([]);
  const log = (text: string) => {
    console.log(text);
    setDebugLogs((logs) => [...logs, text]);
  };

  // useEffect(() => {
  //   (async () => {
  //     const logs = await Updates.readLogEntriesAsync();
  //     for (const l of logs) {
  //       log(JSON.stringify(l));
  //     }
  //   })();
  // }, []);

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
    <View f={1}>
      <Button
        onPress={async () => {
          try {
            const res = await Updates.checkForUpdateAsync();
            log(`checkForUpdateAsync success res=${JSON.stringify(res)}`);
          } catch (e) {
            log(`checkForUpdateAsync error err=${e.message}`);
          }
        }}
      >
        <Text>Check manually for updates</Text>
      </Button>
      {showDownloadButton ? (
        <Button
          onPress={async () => {
            try {
              const res = await Updates.fetchUpdateAsync();
              log(`fetchUpdateAsync success res=${JSON.stringify(res)}`);
            } catch (e) {
              log(`fetchUpdateAsync error err=${e.message}`);
            }
          }}
        >
          <Text>Download and run update</Text>
        </Button>
      ) : null}
      <StatusBar style="auto" />
      <ScrollView f={1}>
        <Text>Updates.channel: {JSON.stringify(Updates.channel)}</Text>
        <Text>
          Updates.checkAutomatically:{" "}
          {JSON.stringify(Updates.checkAutomatically)}
        </Text>
        <Text>Updates.createdAt: {JSON.stringify(Updates.createdAt)}</Text>
        <Text>
          Updates.emergencyLaunchReason:{" "}
          {JSON.stringify(Updates.emergencyLaunchReason)}
        </Text>
        <Text>
          Updates.isEmbeddedLaunch: {JSON.stringify(Updates.isEmbeddedLaunch)}
        </Text>
        <Text>
          Updates.isEmergencyLaunch: {JSON.stringify(Updates.isEmergencyLaunch)}
        </Text>
        <Text>Updates.isEnabled: {JSON.stringify(Updates.isEnabled)}</Text>
        <Text>Updates Demo!</Text>
        <Text>{runTypeMessage}</Text>
        {debugLogs.reverse().map((log, i) => (
          <Text key={i}>{log}</Text>
        ))}
      </ScrollView>
    </View>
  );
}
