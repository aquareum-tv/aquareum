import { useEffect, useState } from "react";
import {
  Text,
  View,
  YStack,
  styled,
  XStack,
  useMedia,
  useWindowDimensions,
  H1,
} from "tamagui";
import * as chrono from "chrono-node";

const CountdownBox = styled(View, {
  alignSelf: "center",
  flexDirection: "row",
  variants: {
    small: {
      true: {
        alignSelf: "auto",
      },
    },
  } as const,
});
const Line = styled(View, {
  // alignItems: "center",
  // flexWrap: "wrap",
  justifyContent: "flex-end",
  flexDirection: "row",
  variants: {
    small: {
      true: {
        flex: 0,
        flexDirection: "column",
      },
    },
  },
} as const);
const Unit = styled(YStack, {
  margin: "$4",
  flex: 0,
} as const);
const BorderBox = styled(View, {
  borderColor: "white",
  borderWidth: 0,
  borderBlockStyle: "solid",
  borderTopWidth: 4,
  // backgroundColor: "green",
} as const);
const TimeText = styled(Text, {
  fontFamily: "$mono",
  fontSize: "$10",
  // backgroundColor: "red",
  // position: "relative",
  // top: 5,
  lineHeight: "$10",
} as const);
const LabelText = styled(Text, {
  fontSize: "$7",
});

const LabelBox = ({ children }) => {
  return (
    <BorderBox>
      <LabelText>{children}</LabelText>
    </BorderBox>
  );
};

export function Countdown({ from, to }: { from?: string; to?: string } = {}) {
  const media = useMedia();
  const [now, setNow] = useState(Date.now());
  const [dest, setDest] = useState<number | null>(null);
  const { width, height } = useWindowDimensions();
  useEffect(() => {
    if (from) {
      const fromDate = chrono.parseDate(from);
      if (fromDate === null) {
        throw new Error("could not parse from");
      }
      setDest(fromDate.getTime());
    } else if (to) {
      const toDate = chrono.parseDate(to);
      if (toDate === null) {
        throw new Error("could not parse to");
      }
      setDest(toDate.getTime());
    } else {
      throw new Error("must provide either from or to");
    }
  }, [from, to]);
  useEffect(() => {
    const tick = () => {
      if (!running) {
        return;
      }
      requestAnimationFrame(tick);
      setNow(Date.now());
    };
    let running = true;
    tick();
    return () => {
      running = false;
    };
  }, []);

  if (dest === null) {
    return <View />;
  }
  let diff = Math.abs(dest - now);
  if (to && now > dest) {
    diff = 0;
  } else if (from && now < dest) {
    diff = 0;
  }
  const small = width <= 600;
  const [years, days, hrs, min, sec, ms] = toLabels(diff);

  return (
    <CountdownBox small={small}>
      <Line small={small}>
        <Unit>
          <TimeText>{years}</TimeText>
          <LabelBox>YEARS</LabelBox>
        </Unit>
        <Unit>
          <TimeText>{days}</TimeText>
          <LabelBox>DAYS</LabelBox>
        </Unit>
        <Unit>
          <TimeText>{hrs}</TimeText>
          <LabelBox>HRS</LabelBox>
        </Unit>
      </Line>
      <Line small={small}>
        <Unit>
          <TimeText>{min}</TimeText>
          <LabelBox>MIN</LabelBox>
        </Unit>
        <Unit>
          <TimeText>{sec}</TimeText>
          <LabelBox>SEC</LabelBox>
        </Unit>
        <Unit>
          <TimeText>{ms}</TimeText>
          <LabelBox>MS</LabelBox>
        </Unit>
      </Line>
    </CountdownBox>
  );
}

const toLabels = (
  now: number,
): [string, string, string, string, string, string] => {
  const ms = now % 1000;
  now = Math.floor(now / 1000);

  const sec = now % 60;
  now = Math.floor(now / 60);

  const min = now % 60;
  now = Math.floor(now / 60);

  const hrs = now % 24;
  now = Math.floor(now / 24);

  const days = now % 365;
  now = Math.floor(now / 365);

  const years = now;

  return [
    pad(years, 4),
    pad(days, 3),
    pad(hrs, 2),
    pad(min, 2),
    pad(sec, 2),
    pad(ms, 3),
  ];
};

const pad = (num: number, n: number): string => {
  let str = `${num}`;
  while (str.length < n) {
    str = "0" + str;
  }
  return str;
};
