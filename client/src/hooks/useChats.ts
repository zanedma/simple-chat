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

/**
 * @description manages the current chat state and authenticating and communicating with the server
 * @returns various state variables and functions to use 
 */
export function useChats() {
  const [chats, setChats] = useState<Record<string, ChatMessage>>({});
  const [username, setUsername] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState("");
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // make sure that if the socket does not exist isConnected is false
    if (!socketRef.current) {
      setIsConnected(false)
    }
  }, [socketRef])

  /**
   * @description exchange a password for an authentication token on the backend
   * @param password password to authenticate with
   * @returns the token if successful, throws and error if not
   */
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

  /**
   * @description initialize a socket connection with the server
   * @param token authentication token received from the server to authorize the connection
   */
  const initSocket = useCallback((token: string) => {
    const url = new URL(BASE_URL);
    url.protocol = "ws";
    url.pathname = "/chat";
    url.searchParams.append("token", token);
    const socket = new WebSocket(url.toString());
    socket.onopen = () => {
      console.log("socket opened");
      setIsConnected(true);
    };
    socket.onerror = () => {
      // unfortunately, the event received does not contain the error that occurred
      // it is assumed that onclose will fire, and that event will contain a code
      // indicating what the error could have been
      console.log("socket error");
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

  /**
   * @description connect to the server by first attempting to get an access token
   * and if successful, using the token establish a socket connection
   * @param newUsername username entered by the user
   * @param password password entered by the user
   */
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

  /**
   * @description send a chat to the server to be saved and broadcasted to other clients
   * @param chatMessage chat message to send
   * @returns the id of the created chat if successful, null if an error occurred
   */
  const sendChat = async (chatMessage: string): Promise<string | null> => {
    if (!socketRef.current) {
      // this should never happen
      setIsConnected(false);
      setError("Attempt to send chat when socket is null (disconnected)");
      return null;
    }
    // chat id is random uuid for simplicity - could also hash the username, chat message, and timestamp
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
    return chatId
  };

  /**
   * @description disconnect the socket
   */
  const disconnect = (): void => {
    socketRef.current?.close();
  };

  return { isConnected, chats, connect, error, username, disconnect, sendChat };
}
