import {
  Box,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  FormLabel,
  Input,
  Button,
  Center,
  Text,
  ModalBody,
} from "@chakra-ui/react";
import { useState } from "react";

interface SignInModalProps {
  connect: (username: string, password: string) => Promise<void>;
  error: string;
}

export default function SignInModal({ connect, error }: SignInModalProps) {
  const { onClose } = useDisclosure();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (
    e: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    e.preventDefault();
    setIsLoading(true)
    await connect(username, password);
    setIsLoading(false)
  };

  return (
    <Box>
      <Modal isOpen={true} onClose={onClose}>
        <ModalOverlay />
        <ModalContent padding={2}>
          <ModalHeader>Sign in</ModalHeader>
          <ModalBody>
            <form>
              <FormLabel marginY={2}>Username</FormLabel>
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
              <FormLabel marginY={2}>Password</FormLabel>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <Center>{error && <Text color="red">{error}</Text>}</Center>
              <Center>
                <Button
                  type="submit"
                  onClick={handleSubmit}
                  marginY={2}
                  colorScheme="blue"
                  isDisabled={!username || !password}
                  isLoading={isLoading}
                >
                  Enter
                </Button>
              </Center>
            </form>
          </ModalBody>
        </ModalContent>
      </Modal>
    </Box>
  );
}
