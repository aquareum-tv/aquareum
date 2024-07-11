import background from "./background";
import "../tamagui-web.css";
import { Link } from "expo-router";
import { Anchor, Button, useTheme, Text, styled, isWeb } from "tamagui";

import { useEffect } from "react";
import { useColorScheme } from "hooks/useColorScheme";
import {
  DarkTheme,
  DefaultTheme,
  ThemeProvider,
} from "@react-navigation/native";
import { useFonts } from "expo-font";
import { SplashScreen, Stack } from "expo-router";
import { Provider } from "./Provider";
import "./updates";
import { Helmet } from "react-native-helmet-async";

export {
  // Catch any errors thrown by the Layout component.
  ErrorBoundary,
} from "expo-router";

export const unstable_settings = {
  // Ensure that reloading on `/modal` keeps a back button present.
  initialRouteName: "(tabs)",
};

// Prevent the splash screen from auto-hiding before asset loading is complete.
SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [fontLoaded, fontError] = useFonts({
    "FiraCode-Light": require("../assets/fonts/FiraCode-Light.ttf"),
    "FiraCode-Medium": require("../assets/fonts/FiraCode-Medium.ttf"),
    "FiraCode-Bold": require("../assets/fonts/FiraCode-Bold.ttf"),
    "FiraSans-Medium": require("../assets/fonts/FiraSans-Medium.ttf"),
  });

  useEffect(() => {
    if (fontLoaded || fontError) {
      // Hide the splash screen after the fonts have loaded (or an error was returned) and the UI is ready.
      SplashScreen.hideAsync();
    }
  }, [fontLoaded, fontError]);

  if (!fontLoaded && !fontError) {
    return null;
  }

  return <RootLayoutNav />;
}

export const LinkNoUnderline = styled(Link, {});

function RootLayoutNav() {
  const colorScheme = useColorScheme();

  useEffect(() => {
    background();
  }, [])

  return (
    <Provider>
      <ThemeProvider value={colorScheme === "dark" ? DarkTheme : DefaultTheme}>
        {isWeb && (
          <Helmet>
            <title>Aquareum</title>
          </Helmet>
        )}
        <Stack>
          <Stack.Screen
            name="(tabs)"
            options={{
              title: "",
              headerShown: true,
              headerRight: () => (
                <Anchor href="https://explorer.livepeer.org/treasury/74518185892381909671177921640414850443801430499809418110611019961553289709442">
                  <Button
                    mr="$4"
                    bg="rgb(189 110 134)"
                    color="white"
                    fontSize="$6"
                  >
                    What's Aquareum?
                  </Button>
                </Anchor>
              ),
            }}
          />

          {/* <Stack.Screen
            name="about"
            options={{
              title: "What's Aquareum?",
              presentation: "modal",
              animation: "slide_from_right",
              gestureEnabled: true,
              gestureDirection: "horizontal",
              headerShown: true,
            }}
          /> */}
        </Stack>
      </ThemeProvider>
    </Provider>
  );
}
