import React from "react";
import SharedProvider from "./provider.shared";

export default function Provider({ children }: { children: React.ReactNode }) {
  return <SharedProvider>{children}</SharedProvider>;
}
