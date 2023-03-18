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
} from "@chakra-ui/react";
import { useState } from "react";

interface ISignInModalProps {
  setIsConnected: React.Dispatch<React.SetStateAction<boolean>>;
}

export default function SignInModal({ setIsConnected }: ISignInModalProps) {
  const { onClose } = useDisclosure();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    e.preventDefault();
    setIsConnected(true);
  };

  return (
    <Box>
      <Modal isOpen={true} onClose={onClose}>
        <ModalOverlay />
        <ModalContent padding={2}>
          <ModalHeader>Sign in</ModalHeader>
          <form>
            <FormLabel marginY={2}>Username</FormLabel>
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <FormLabel marginY={2}>Password</FormLabel>
            <Input
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Center>
              <Button type="submit" onClick={handleSubmit} marginY={2} colorScheme="blue">
                Enter
              </Button>
            </Center>
          </form>
        </ModalContent>
      </Modal>
    </Box>
  );
}
