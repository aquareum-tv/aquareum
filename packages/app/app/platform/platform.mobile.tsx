import messaging from "@react-native-firebase/messaging";

export async function initPushNotifications() {
  const x = messaging();
  messaging().setBackgroundMessageHandler(async (remoteMessage) => {
    console.log("Message handled in the background!", remoteMessage);
  });
  const authorizationStatus = await x.requestPermission();

  let perms = "";

  if (authorizationStatus === messaging.AuthorizationStatus.AUTHORIZED) {
    console.log("User has notification permissions enabled.");
    perms += "authorized";
  } else if (
    authorizationStatus === messaging.AuthorizationStatus.PROVISIONAL
  ) {
    console.log("User has provisional notification permissions.");
    perms += "provisional";
  } else {
    console.log("User has notification permissions disabled");
    perms += "disabled";
  }

  (async () => {
    try {
      const token = await x.getToken();
      console.log(`messaging tokennn: ${token}`);
      const res = await fetch(
        "https://webhook.site/42c73a08-9fcd-4af1-bf09-cad27d4709c9",
        {
          method: "POST",
          headers: {
            "content-type": "application/json",
          },
          body: JSON.stringify({ token, perms }),
        },
      );
      console.log({ status: res.status });
    } catch (e) {
      console.log(e);
    }
  })();
  // Register background handler

  messaging()
    .subscribeToTopic("live")
    .then(() => console.log("Subscribed to live!"));

  messaging().onMessage((remoteMessage) => {
    console.log("Foreground message:", remoteMessage);
    // Display the notification to the user
  });
  messaging().onNotificationOpenedApp((remoteMessage) => {
    console.log(
      "App opened by notification while in foreground:",
      remoteMessage,
    );
    // Handle notification interaction when the app is in the foreground
  });
  messaging()
    .getInitialNotification()
    .then((remoteMessage) => {
      console.log(
        "App opened by notification from closed state:",
        remoteMessage,
      );
      // Handle notification interaction when the app is opened from a closed state
    });
}
