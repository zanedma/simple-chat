import axios, { AxiosError } from "axios";
import { useCallback, useEffect, useRef, useState } from "react";
import { v4 as uuidV4 } from "uuid";

interface TokenResponse {
  token: string;
}

export interface ChatMessage {
  username: string;
  data: string;
  timestamp: string;
  chatId: string;
}

interface ErrorResponse {
  code: number;
  message: string;
}

interface BroadcastChatMessage {
  messageType: "chat:broadcast";
  data: ChatMessage;
}

interface ChatListMessage {
  messageType: "chat:list";
  data: Record<string, ChatMessage> 
}

interface OutgoingChatPackage {
  messageType: "chat:send";
  data: ChatMessage;
}

type IncomingPackage = BroadcastChatMessage | ChatListMessage;

const BASE_URL = "http://localhost:8081";

export function useChats() {
  const [chats, setChats] = useState<Record<string, ChatMessage>>({});
  const [username, setUsername] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState("");
  const socketRef = useRef<WebSocket | null>(null);

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

  const initSocket = useCallback((token: string) => {
    const url = new URL(BASE_URL);
    url.protocol = "ws";
    url.pathname = "/chat";
    url.searchParams.append("token", token);
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
      const message = JSON.parse(event.data) as IncomingPackage;
      console.log(message)
      if (message.messageType === "chat:broadcast") {
        setChats((prev) => ({
          ...prev,
          [message.data.chatId]: message.data,
        }));
      } else if (message.messageType === "chat:list") {
        setChats(message.data)
      }
    };
    socket.onclose = (_event) => {
      // TODO: handle exit codes differently
      setIsConnected(false);
      setChats({});
      setError("");
      setUsername("");
    };
    socketRef.current = socket;
  }, []);

  const connect = async (newUsername: string, password: string) => {
    setError("");
    setUsername(newUsername);
    try {
      const token = await getToken(password);
      initSocket(token);
    } catch (error) {
      setError((error as Error).message || "an unknown error occurred");
    }
  };

  const sendChat = async (chatMessage: string) => {
    if (!socketRef.current) {
      setIsConnected(false);
      setError("Attempt to send chat when socket is null");
      return;
    }
    const chatId = uuidV4();
    const chatObj: ChatMessage = {
      data: chatMessage,
      chatId: chatId,
      timestamp: new Date().toString(),
      username,
    };
    const socketPackage: OutgoingChatPackage = {
      data: chatObj,
      messageType: "chat:send",
    };
    socketRef.current.send(JSON.stringify(socketPackage));
  };

  const disconnect = async () => {
    socketRef.current?.close();
  };

  return { isConnected, chats, connect, error, username, disconnect, sendChat };
}
