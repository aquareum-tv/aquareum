import { View, isWeb } from "tamagui";
import { useEffect } from "react";

export default function SupportScreen() {
  if (isWeb) {
    useEffect(() => {
      document.location.href =
        "https://docs.google.com/forms/d/14ATDKwOkSN1SDxb_anMT1iafs3JtyXSoubSBEoJuA5g/edit";
    }, []);
  }
  return <View />;
}
