import { Link, Tabs } from "expo-router";
import { Button, useTheme, View } from "tamagui";
import { Atom, AudioWaveform } from "@tamagui/lucide-icons";

export default function TabLayout() {
  const theme = useTheme();

  return (
    // <MainScreen />
    <Tabs
      screenOptions={{
        tabBarActiveTintColor: theme.red10.val,
      }}
      tabBar={() => <View></View>}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: "Aquareum",
          tabBarIcon: ({ color }) => <Atom color={color} />,
          headerShown: false,
        }}
      />
      <Tabs.Screen
        name="admin"
        options={{
          title: "Admin",
          tabBarIcon: ({ color }) => <Atom color={color} />,
          headerShown: false,
        }}
      />
    </Tabs>
  );
}
