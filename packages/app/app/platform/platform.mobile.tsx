import messaging from "@react-native-firebase/messaging";

export async function checkApplicationPermission() {
  const x = messaging();
  messaging().setBackgroundMessageHandler(async (remoteMessage) => {
    console.log("Message handled in the background!", remoteMessage);
  });
  const authorizationStatus = await x.requestPermission();

  if (authorizationStatus === messaging.AuthorizationStatus.AUTHORIZED) {
    console.log("User has notification permissions enabled.");
  } else if (
    authorizationStatus === messaging.AuthorizationStatus.PROVISIONAL
  ) {
    console.log("User has provisional notification permissions.");
  } else {
    console.log("User has notification permissions disabled");
  }
  console.log(await x.getToken());
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
