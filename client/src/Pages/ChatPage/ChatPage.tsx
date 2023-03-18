import { useEffect } from "react";
import { useChats } from "../../hooks/useChats";
import SignInModal from "./SignInModal";

export default function ChatPage() {
  const { isConnected, error, connect } = useChats();
  console.log(isConnected)
  useEffect(() => console.log("3: ", isConnected), [isConnected])
  useEffect(() => console.log("3: ", error), [error])
  useEffect(() => console.log("4: ", connect), [connect])

  return (
    <>
      {!isConnected && <SignInModal />}
      {isConnected && <p>Hello</p>}
    </>
  );
}
