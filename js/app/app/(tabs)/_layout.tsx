import { Link, Tabs } from "expo-router";
import { Button, useTheme } from "tamagui";
import { Atom, AudioWaveform } from "@tamagui/lucide-icons";
import MainScreen from "./index";

export default function TabLayout() {
  const theme = useTheme();

  return (
    <MainScreen />
    // <Tabs
    //   screenOptions={{
    //     tabBarActiveTintColor: theme.red10.val,
    //   }}
    // >
    //   <Tabs.Screen
    //     name="index"
    //     options={{
    //       title: "Aquareum",
    //       tabBarIcon: ({ color }) => <Atom color={color} />,
    //       headerShown: false,
    //     }}
    //   />
    // </Tabs>
  );
}
