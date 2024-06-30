// earliest part of aquareum's entrypoint. set up message notifications and stuff.
// index.js
import messaging from "@react-native-firebase/messaging";
import { checkApplicationPermission } from "./platform";

export default async function background() {
  await checkApplicationPermission();
}
