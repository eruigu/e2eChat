export type ChatCount = {
  type: string;
  count: number;
};

export type ChatMessage = {
  user: string;
  text: string;
  time: string;
};

export type ConnectionStatus = 
  "idle" | 
  "connecting" | 
  "connected" | 
  "disconnected" | 
  "error";

export type IncomingMessage = ChatMessage | ChatCount;
