import React from "react";
import { Text, View } from "react-native";
import { AppRegistry, Platform } from "react-native";

const HelloWorldApp = () => {
  return (
    <View
      style={{
        flex: 1,
        justifyContent: "center",
        alignItems: "center",
      }}
    >
      <Text>Hello, world!</Text>
    </View>
  );
};

AppRegistry.registerComponent("main", () => HelloWorldApp);

if (Platform.OS === "web") {
  const rootTag =
    document.getElementById("root") || document.getElementById("main");
  AppRegistry.runApplication("main", { rootTag });
}
