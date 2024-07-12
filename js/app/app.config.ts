export default function () {
  const pkg = require("./package.json");
  return {
    expo: {
      name: "Aquareum",
      slug: "Aquareum",
      version: pkg.version,
      runtimeVersion: pkg.version,
      orientation: "portrait",
      icon: "./assets/images/icon.png",
      scheme: "myapp",
      userInterfaceStyle: "automatic",
      splash: {
        image: "./assets/images/splash.png",
        resizeMode: "contain",
        backgroundColor: "#ffffff",
      },
      assetBundlePatterns: ["**/*"],
      ios: {
        supportsTablet: true,
        bundleIdentifier: "tv.aquareum",
        googleServicesFile: "./GoogleService-Info.plist",
        entitlements: {
          "aps-environment": "production",
        },
      },
      android: {
        adaptiveIcon: {
          foregroundImage: "./assets/images/adaptive-icon.png",
          backgroundColor: "#ffffff",
        },
        package: "tv.aquareum",
        googleServicesFile: "./google-services.json",
        permissions: [
          "android.permission.SCHEDULE_EXACT_ALARM",
          "android.permission.POST_NOTIFICATIONS",
        ],
      },
      web: {
        bundler: "metro",
        output: "static",
        favicon: "./assets/images/favicon.png",
      },
      plugins: [
        "expo-router",
        [
          "expo-font",
          {
            fonts: [
              "assets/fonts/FiraCode-Bold.ttf",
              "assets/fonts/FiraCode-Light.ttf",
              "assets/fonts/FiraCode-Medium.ttf",
              "assets/fonts/FiraCode-Regular.ttf",
              "assets/fonts/FiraCode-Retina.ttf",
              "assets/fonts/FiraSans-Black.ttf",
              "assets/fonts/FiraSans-BlackItalic.ttf",
              "assets/fonts/FiraSans-Bold.ttf",
              "assets/fonts/FiraSans-BoldItalic.ttf",
              "assets/fonts/FiraSans-ExtraBold.ttf",
              "assets/fonts/FiraSans-ExtraBoldItalic.ttf",
              "assets/fonts/FiraSans-ExtraLight.ttf",
              "assets/fonts/FiraSans-ExtraLightItalic.ttf",
              "assets/fonts/FiraSans-Italic.ttf",
              "assets/fonts/FiraSans-Light.ttf",
              "assets/fonts/FiraSans-LightItalic.ttf",
              "assets/fonts/FiraSans-Medium.ttf",
              "assets/fonts/FiraSans-MediumItalic.ttf",
              "assets/fonts/FiraSans-Regular.ttf",
              "assets/fonts/FiraSans-SemiBold.ttf",
              "assets/fonts/FiraSans-SemiBoldItalic.ttf",
              "assets/fonts/FiraSans-Thin.ttf",
              "assets/fonts/FiraSans-ThinItalic.ttf",
              "assets/fonts/SpaceMono-Regular.ttf",
            ],
          },
        ],
        "@react-native-firebase/app",
        "@react-native-firebase/messaging",
        [
          "expo-build-properties",
          {
            ios: {
              useFrameworks: "static",
            },
          },
        ],
      ],
      experiments: {
        typedRoutes: true,
      },
      updates: {
        url: "https://aquareum.tv/app-updates",
        enabled: true,
        checkAutomatically: "ON_LOAD",
        fallbackToCacheTimeout: 0,
        codeSigningCertificate: "./code-signing/certificate.pem",
        codeSigningMetadata: {
          keyid: "main",
          alg: "rsa-v1_5-sha256",
        },
      },
    },
  };
}
