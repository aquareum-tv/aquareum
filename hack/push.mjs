import admin from "firebase-admin";
import fs from "fs";

const serviceAccount = JSON.parse(
  fs.readFileSync(`${process.env.AQKEYS}/firebase-admin.json`, "utf8"),
);

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
});

(async () => {
  const res = await admin.messaging().sendMulticast({
    tokens: process.argv.slice(2),
    notification: {
      title: "Basic Notification",
      body: "This is a basic notification sent from the server!",
      // imageUrl: "https://my-cdn.com/app-logo.png",
    },
    android: {
      priority: "high",
    },
  });
  console.log(JSON.stringify(res));
})();
