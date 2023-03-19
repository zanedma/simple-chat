import { Box, Text } from "@chakra-ui/react";
import { ChatMessage } from "../../hooks/useChats";

interface MessageProps {
  username: string;
  message: ChatMessage;
  isPending: boolean;
}

export default function ChatMessageBox({
  username,
  message,
  isPending,
}: MessageProps) {
  const isUsersMessage = message.username === username;
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
          {message.username}
        </Text>
      )}
      <Text>{message.data}</Text>
      <Text
        fontSize="xs"
        color="blackAlpha.500"
        marginLeft="auto"
        marginRight={0}
      >
        {isPending ? "Sending..." : date.toString()}
      </Text>
    </Box>
  );
}
