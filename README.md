# Aquareum

## The Video Layer for Everything

You'll need Node, Yarn, Meson, Ninja, and Go.

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

**Web Development**

```
yarn run app start
```

**App Development**

```
cd js/app

# iOS
npx expo run:ios -d "Your iPhone"

# Android
npx expo run:android

# Web http://localhost:38081
npx expo start
```

You can also boot the other platforms directly from the bundler once it starts.
