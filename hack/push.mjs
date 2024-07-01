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
    tokens: [
      "dGCfL2MtzkeQpYqrTCiYPG:APA91bEZBZwkZER7s6nF8FifUfm3I57NwCiOAKPtfuJdyEypaGnDFgIJMWYUo4w15-ODQoy0U3u8XOgM5kxxVrf21BPsq_SPsvPzmw3pljPK4oCed3SDHEtGqH2qA86FR8OIbK61dnlj",
    ],
    notification: {
      title: "Basic Notification",
      body: "This is a basic notification sent from the server!",
      // imageUrl: "https://my-cdn.com/app-logo.png",
    },
  });
  console.log(JSON.stringify(res));
})();
