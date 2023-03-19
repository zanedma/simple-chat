import { Box, Text } from "@chakra-ui/react";
import { ChatMessage } from "../../types";

interface MessageProps {
  username: string;
  message: ChatMessage;
}
export default function ChatMessageBox({ username, message }: MessageProps) {
  const isUsersMessage = message.clientId === username;
  const date = new Date(message.timestamp);
  return (
    <Box
      marginY={4}
      marginLeft={isUsersMessage ? "auto" : 4}
      marginRight={isUsersMessage ? 4 : "auto"}
      background={isUsersMessage ? "blue.200" : "gray.200"}
      borderRadius="md"
      shadow="md"
      maxWidth="45%"
      padding={2}
    >
      {!isUsersMessage && (
        <Text fontSize="xs" color="blackAlpha.500">
          {message.clientId}
        </Text>
      )}
      <Text>{message.message}</Text>
      <Text
        fontSize="xs"
        color="blackAlpha.500"
        marginLeft="auto"
        marginRight={0}
      >
        {date.toString()}
      </Text>
    </Box>
  );
}
