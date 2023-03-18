import { useState } from "react";
import SignInModal from "./SignInModal";

export default function ChatPage() {
  const [isConnected, setIsConnected] = useState(false)

  if (!isConnected) {
    return <SignInModal setIsConnected={setIsConnected} />
  }

  return (
    <>Authenticated</>
  )
}
