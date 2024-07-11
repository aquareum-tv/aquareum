// earliest part of aquareum's entrypoint. set up message notifications and stuff.
// index.js
import { initPushNotifications } from "./platform";

export default async function background() {
  await initPushNotifications();
}
