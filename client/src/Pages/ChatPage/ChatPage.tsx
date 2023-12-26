import {
  Box,
  Button,
  Center,
  Divider,
  Flex,
  Spacer,
  Text,
  Textarea,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { ChatMessage, useChats } from "../../hooks/useChats";
import ChatMessageBox from "./ChatMessageBox";
import SignInModal from "./SignInModal";

export default function ChatPage() {
  const { isConnected, error, connect, username, disconnect, chats, sendChat } =
    useChats();
  const [messageInput, setMessageInput] = useState("");
  const [sendError, setSendError] = useState("");
  const [pendingId, setPendingId] = useState("")
  const [sendTimeout, setSendTimeout] = useState<NodeJS.Timeout | null>(null)

  useEffect(() => {
    if (pendingId && chats[pendingId]) {
      setPendingId("")
      setMessageInput("")
      if (sendTimeout) {
        clearTimeout(sendTimeout)
        setSendTimeout(null)
      }
    }
  }, [chats, pendingId, sendTimeout])

  const handleSend = async () => {
    const messageId = await sendChat(messageInput);
    if (!messageId) {
      setSendError("Error sending message");
    } else {
      setPendingId(messageId)
      // if the message times out, disconnect the connection for simplicitys sake
      // there are better ways to handle this
      setSendTimeout(setTimeout(() => {
        setPendingId("")
        disconnect()
      }, 5000))
    }
  };

  const sortByTimeStamp = (a: ChatMessage, b: ChatMessage) => {
    return new Date(a.timestamp).valueOf() - new Date(b.timestamp).valueOf();
  };

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
            Simple Chat
          </Text>
          <Spacer />
          <Text fontSize="xl">{username}</Text>
          <Button size="xs" colorScheme="blue" onClick={disconnect}>
            Disconnect
          </Button>
        </Flex>
        <Divider marginY={2} />
        {Object.values(chats)
          .sort(sortByTimeStamp)
          .map((message, idx) => (
            <ChatMessageBox username={username} message={message} key={idx} isPending={false} />
          ))}
        <form>
          <Textarea
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            isDisabled={!!pendingId}
          />
          {sendError && <Text color="red">{sendError}</Text>}
          <Center marginY={2}>
            <Button
              colorScheme="blue"
              onClick={handleSend}
              isDisabled={!messageInput}
              isLoading={!!pendingId}
            >
              Send
            </Button>
          </Center>
        </form>
      </Box>
    </Center>
  );
}
