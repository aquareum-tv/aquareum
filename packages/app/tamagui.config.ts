import { config as configBase } from "@tamagui/config/v3";
import { createTamagui, createFont } from "tamagui";

const bodyFont = createFont({
  ...configBase.fonts.body,
  family: `FiraSans-Medium`,
});

const headingFont = createFont({
  ...configBase.fonts.heading,
  family: `FiraSans-Medium`,
});

const codeFont = createFont({
  family: `FiraCode-Medium`,
  size: {
    1: 12,
    2: 14,
    3: 15,
    4: 16,
  },
  lineHeight: {
    // 1 will be 22
    2: 22,
  },
  weight: {
    1: "300",
    // 2 will be 300
    3: "600",
  },
  letterSpacing: {
    1: 0,
    2: -1,
    // 3 will be -1
  },
});

const aquareumConfig = {
  ...configBase,
  fonts: {
    ...configBase.fonts,
    heading: headingFont,
    body: bodyFont,
    mono: codeFont,
  },
};

const config = createTamagui(aquareumConfig);

export { config };
export default config;

export type Conf = typeof config;

declare module "tamagui" {
  interface TamaguiCustomConfig extends Conf {}
}
