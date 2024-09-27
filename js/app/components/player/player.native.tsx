import React from "react";
import WebView from "react-native-webview";
import { View, Text } from "tamagui";

export function Player() {
  return (
    <View f={1}>
      <WebView
        allowsInlineMediaPlayback={true}
        scrollEnabled={false}
        source={{
          uri: "http://127.0.0.1:38081/embed/0x6fbe6863cf1efc713899455e526a13239d371175",
        }}
        style={{ flex: 1, backgroundColor: "green" }}
      ></WebView>
    </View>
  );
}
