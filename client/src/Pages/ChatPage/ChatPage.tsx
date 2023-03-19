import {
  Box,
  Button,
  Center,
  Divider,
  Flex,
  Spacer,
  Text,
} from "@chakra-ui/react";
import { useChats } from "../../hooks/useChats";
import ChatMessageBox from "./ChatMessageBox";
import SignInModal from "./SignInModal";

export default function ChatPage() {
  const { isConnected, error, connect, username, disconnect, chats } =
    useChats();

  if (!isConnected) {
    return <SignInModal connect={connect} error={error} />;
  }

  return (
    <Center marginTop="10vh">
      <Box
        width="80vw"
        padding={4}
        borderColor="blackAlpha.400"
        borderWidth="1px"
        borderRadius="md"
        shadow="md"
        marginBottom={8}
      >
        <Flex alignItems="baseline" gap={4}>
          <Text fontSize="2xl" fontWeight="semibold">
            Beehive Chat
          </Text>
          <Spacer />
          <Text fontSize="xl">{username}</Text>
          <Button size="xs" colorScheme="blue" onClick={disconnect}>
            Disconnect
          </Button>
        </Flex>
        <Divider marginY={2} />
        {chats.map((message, idx) => (
          <ChatMessageBox username={username} message={message} key={idx} />
        ))}
      </Box>
    </Center>
  );
}
