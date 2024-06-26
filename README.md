#  Aquareum

## The Video Layer for Everything

You'll need Node, Yarn, and Go.

**Single-command build:**
```
make
```

**Dev Setup**
```
yarn install
yarn run build
```

**Node Development**
```
make node
./bin/aquareum
```

**App Development**
```
cd packages/app

# iOS
npx expo run:ios -d "Your iPhone"

# Android
npx expo run:android

# Web http://localhost:8081
npx expo start
```

You can also boot the other platforms directly from the bundler once it starts.
