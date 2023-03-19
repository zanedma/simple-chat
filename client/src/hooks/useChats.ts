import axios, { AxiosError } from "axios";
import { useRef, useState } from "react";
import { ChatMessage } from "../types";

interface TokenResponse {
  token: string;
}

interface ErrorResponse {
  code: number;
  message: string;
}

interface BroadcastChatMessage {
  messageType: "chat:broadcast";
  data: ChatMessage;
}

type IncomingMessage = BroadcastChatMessage;

const BASE_URL = "http://localhost:8081";
const tempChats: ChatMessage[] = [
  {clientId: "user1", message: "Hello there", timestamp: new Date().toString(), messageId: "1"},
  {clientId: "zane", message: "Hi", timestamp: new Date().toString(), messageId: "2"},
  {clientId: "zane", message: "Hi", timestamp: new Date().toString(), messageId: "3"},
  {clientId: "zane", message: "Hi", timestamp: new Date().toString(), messageId: "4"},
]

export function useChats() {
  const [chats, setChats] = useState<ChatMessage[]>(tempChats);
  const [username, setUsername] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState("");
  const socketRef = useRef<WebSocket | null>(null)

  const getToken = async (password: string): Promise<string> => {
    try {
      const res = await axios.get<TokenResponse>(`${BASE_URL}/auth`, {
        headers: {
          "X-Connection-Password": password,
        },
      });
      return res.data.token;
    } catch (error) {
      const axErr = error as AxiosError;
      if (axErr.isAxiosError) {
        throw new Error(
          (axErr.response?.data as ErrorResponse).message ||
            axErr.message ||
            "an unknown error occurred"
        );
      } else {
        throw error;
      }
    }
  };

  const initSocket = (token: string) => {
    const url = new URL(BASE_URL);
    url.protocol = "ws";
    url.pathname = "/chat";
    url.searchParams.append("token", token);
    console.log(url.toString());
    const socket = new WebSocket(url.toString());
    socket.onopen = (event) => {
      console.log("socket opened");
      setIsConnected(true);
    };
    socket.onerror = (event) => {
      console.log("socket error");
      console.log(event);
    };
    socket.onmessage = (event) => {
      console.log("socket message");
      console.log(event.data);
      const message = event.data as IncomingMessage;
      if (message.messageType === "chat:broadcast") {
        setChats([...chats, message.data]);
      }
    };
    socket.onclose = (event) => {
      // TODO: handle exit codes differently
      setIsConnected(false)
      setChats([]);
      setError("");
      setUsername("")
    };
    socketRef.current = socket
  };

  const connect = async (newUsername: string, password: string) => {
    setError("");
    setUsername(newUsername)
    try {
      const token = await getToken(password);
      console.log("successfully got token");
      initSocket(token);
    } catch (error) {
      console.log(error);
      setError((error as Error).message || "an unknown error occurred");
    }
  };

  const disconnect = async () => {
    socketRef.current?.close()
  }

  return { isConnected, chats, connect, error, username, disconnect };
}
