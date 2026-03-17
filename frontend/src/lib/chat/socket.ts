import { addMessage, errorMessage, onlineCount, status } from "./store";
import type { IncomingMessage } from "./types";

export class ChatSocket {
  private socket: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private manuallyClosed = false;

  connect(name: string){
    if (this.socket && this.socket.readyState === WebSocket.OPEN) return;

    this.manuallyClosed = false;
    status.set("connecting");
    errorMessage.set("");

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const host = "localhost:8080";
    const url = `${protocol}://${host}/subscribe?name=${encodeURIComponent(name)}`;

    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      this.reconnectAttempts = 0;
      status.set("connected")
    };

    this.socket.onmessage = (e) => {
      try {
        const data: IncomingMessage = JSON.parse(e.data);
        console.log(data)

        if ("type" in data && data.type === "online_count"){
          onlineCount.set(data.count)
          return
        }

        addMessage(data);
      } catch {
        errorMessage.set("Received invalid message payload")
      }
    };

    this.socket.onclose = () => {
      this.socket = null;

      if(this.manuallyClosed){
        status.set("disconnected")
        return;
      }

      status.set("disconnected")
      this.scheduleReconnect(name);
    };
  }

  send(message: ChatMessage){
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return false;

    this.socket.send(JSON.stringify(message));
    return true;
  }

  disconnect(){
    this.manuallyClosed = true;

    if (this.reconnectTimer){
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null;
    }

    this.socket?.close()
    this.socket = null;
    status.set("disconnected")
  }
  private scheduleReconnect(name: string) {
    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 10000);
    this.reconnectAttempts += 1;

    this.reconnectTimer = setTimeout(() => {
      this.connect(name);
    }, delay);
  }
}

export const chatSocket = new ChatSocket();
