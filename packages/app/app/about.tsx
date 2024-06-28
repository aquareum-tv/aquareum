import {
  Paragraph,
  ScrollView,
  View,
  Text,
  styled,
  H1,
  H2,
  H3,
  H4,
  H5,
  H6,
  isWeb,
} from "tamagui";
import { Asset } from "expo-asset";
import { useEffect, useState } from "react";
import Markdown from "react-native-markdown-display";
import { description } from "./aquareum-description";
import { SafeAreaView } from "react-native";
import { Dimensions } from "react-native";

const Code = styled(Text, { fontFamily: "$mono" });

const rules = {
  // Headings
  heading1: (node, children, parent, styles) => (
    <H1
      key={node.key}
      style={styles._VIEW_SAFE_heading1}
      paddingVertical="$4"
      lineHeight="$9"
    >
      {children}
    </H1>
  ),
  heading2: (node, children, parent, styles) => (
    <H2 key={node.key} style={styles._VIEW_SAFE_heading2}>
      {children}
    </H2>
  ),
  heading3: (node, children, parent, styles) => (
    <H3 key={node.key} style={styles._VIEW_SAFE_heading3}>
      {children}
    </H3>
  ),
  heading4: (node, children, parent, styles) => (
    <H4 key={node.key} style={styles._VIEW_SAFE_heading4}>
      {children}
    </H4>
  ),
  heading5: (node, children, parent, styles) => (
    <H5 key={node.key} style={styles._VIEW_SAFE_heading5}>
      {children}
    </H5>
  ),
  heading6: (node, children, parent, styles) => (
    <H6 key={node.key} style={styles._VIEW_SAFE_heading6}>
      {children}
    </H6>
  ),
  text: (node, children, parent, styles, inheritedStyles = {}) => (
    <Text
      lineHeight="$6"
      fontSize="$7"
      key={node.key}
      style={[inheritedStyles, styles.text]}
    >
      {node.content}
    </Text>
  ),
  paragraph: (node, children, parent, styles) => (
    <Paragraph key={node.key} style={styles._VIEW_SAFE_paragraph}>
      {children}
    </Paragraph>
  ),
  bullet_list: (node, children, parent, styles) => (
    <View key={node.key} style={styles._VIEW_SAFE_bullet_list}>
      {children}
    </View>
  ),
  ordered_list: (node, children, parent, styles) => (
    <View key={node.key} style={styles._VIEW_SAFE_ordered_list}>
      {children}
    </View>
  ),
  code_inline: (node, children, parent, styles, inheritedStyles = {}) => (
    <Code lineHeight="$6" fontSize="$7" key={node.key}>
      {node.content}
    </Code>
  ),
  pre: (node, children, parent, styles) => (
    <View key={node.key} style={styles._VIEW_SAFE_pre}>
      {children}
    </View>
  ),
  list_item: (node, children, parent, styles) => (
    <Paragraph key={node.key} style={styles._VIEW_SAFE_paragraph}>
      <Text>&gt;</Text> {children}
    </Paragraph>
  ),
};

export default function ModalScreen() {
  if (isWeb) {
    return (
      <ScrollView
        backgroundColor="$background"
        padding="$4"
        borderRadius="$4"
        height="100%"
      >
        <View
          paddingBottom={30}
          flex={1}
          maxWidth="800px"
          alignSelf={isWeb ? "center" : "stretch"}
        >
          <Markdown rules={rules}>{description}</Markdown>
        </View>
      </ScrollView>
    );
  }
  return (
    <SafeAreaView>
      <ScrollView
        backgroundColor="$background"
        padding="$4"
        borderRadius="$4"
        height="100%"
      >
        <View paddingBottom={30} flex={1}>
          <Markdown rules={rules}>{description}</Markdown>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
