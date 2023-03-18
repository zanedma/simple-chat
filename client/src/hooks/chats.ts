import { useState } from "react";
import { Message } from "../types";
import WebSocket from "ws"

export function useChats() {
  const [chats, setChats] = useState<Message[]>([])
  const socket = new WebSocket("ws://localhost:8081", {headers: {}})
}