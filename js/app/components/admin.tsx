import {
  Button,
  View,
  TextArea,
  Input,
  Label,
  H5,
  Text,
  Paragraph,
} from "tamagui";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { useSignTypedData, useAccount } from "wagmi";
import schema from "generated/eip712-schema.json";
import { useState } from "react";
import { useToastController } from "@tamagui/toast";
import { EXPO_PUBLIC_AQUAREUM_URL } from "constants/env";

export default function AdminPage() {
  const { signTypedDataAsync } = useSignTypedData();
  const account = useAccount();
  const [streamer, setStreamer] = useState("");
  const [title, setTitle] = useState("");
  const [loading, setLoading] = useState(false);
  const toast = useToastController();
  const [streamKey, setStreamKey] = useState("");
  const disabled = loading || streamer === "" || title === "";
  return (
    <View f={1} ai="center" jc="center">
      <ConnectButton />
      {account.address && (
        <View>
          <Label>
            Streamer
            <Input value={streamer} onChangeText={setStreamer} />
          </Label>
          <Label>
            Message
            <TextArea value={title} onChangeText={setTitle} />
          </Label>
          <Button
            disabled={disabled}
            opacity={disabled ? 0.5 : 1}
            onPress={async () => {
              try {
                setLoading(true);
                const message = {
                  signer: account.address,
                  time: Date.now(),
                  data: { streamer, title },
                };
                const signature = await signTypedDataAsync({
                  types: schema.types,
                  domain: schema.domain,
                  primaryType: "GoLive",
                  message: message,
                });
                const res = await fetch(
                  `${EXPO_PUBLIC_AQUAREUM_URL}/api/golive`,
                  {
                    method: "POST",
                    body: JSON.stringify({
                      primaryType: "GoLive",
                      domain: schema.domain,
                      message: message,
                      signature: signature,
                    }),
                  },
                );
                if (!res.ok) {
                  const text = await res.text();
                  throw new Error(`http ${res.status} ${text}`);
                }
                toast.show("GoLive Succeeded", {
                  message: "Let's goooooo!",
                });
                setStreamer("");
                setTitle("");
              } catch (e) {
                toast.show("GoLive Failed", {
                  message: e.message,
                });
              } finally {
                setLoading(false);
              }
            }}
          >
            {loading ? "Loading..." : "Sign message"}
          </Button>
          <Button
            onPress={async () => {
              try {
                const message = {
                  signer: account.address,
                  time: Date.now(),
                  data: {
                    authorized: "my-server",
                  },
                };
                const signature = await signTypedDataAsync({
                  types: schema.types,
                  domain: schema.domain,
                  primaryType: "StreamKey",
                  message: message,
                });
                let key = JSON.stringify({
                  primaryType: "StreamKey",
                  domain: schema.domain,
                  message: message,
                  signature: signature,
                });
                key = btoa(key);
                key = key.replaceAll("+", "-");
                key = key.replaceAll("/", "_");
                setStreamKey(key);
                toast.show("Created Stream Key", {
                  message: "Let's goooooo!",
                });
                setStreamer("");
                setTitle("");
              } catch (e) {
                toast.show("Stream Key Creation Failed", {
                  message: e.message,
                });
              }
            }}
          >
            {"Generate Stream Key"}
          </Button>
          {streamKey && (
            <View f={1} alignItems="stretch" maxWidth="100vw">
              <H5>Stream Key:</H5>
              <Paragraph p="$10">{streamKey}</Paragraph>
            </View>
          )}
        </View>
      )}
    </View>
  );
}
