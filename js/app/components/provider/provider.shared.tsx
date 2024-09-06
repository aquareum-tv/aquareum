import { ToastProvider, ToastViewport } from "@tamagui/toast";
import { CurrentToast } from "app/CurrentToast";
import React from "react";
import { TamaguiProvider } from "tamagui";
import config from "tamagui.config";
import { AquareumProvider } from "hooks/useAquareumNode";

export default function Provider({ children }: { children: React.ReactNode }) {
  return (
    <AquareumProvider>
      <TamaguiProvider config={config} defaultTheme={"dark"}>
        <ToastProvider
          swipeDirection="vertical"
          duration={6000}
          native={
            [
              /* uncomment the next line to do native toasts on mobile. NOTE: it'll require you making a dev build and won't work with Expo Go */
              // 'mobile'
            ]
          }
        >
          {children}
          <CurrentToast />
          <ToastViewport name="default" top="$8" left={0} right={0} />
        </ToastProvider>
      </TamaguiProvider>
    </AquareumProvider>
  );
}
