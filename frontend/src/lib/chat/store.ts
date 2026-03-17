import { writable } from "svelte/store";
import type { ConnectionStatus, IncomingMessage } from "./types";

export const username = writable("");
export const onlineCount = writable(0);
export const errorMessage = writable("");
export const messages = writable<IncomingMessage[]>([]);
export const status = writable<ConnectionStatus>("idle");

export function addMessage(message: IncomingMessage){
  messages.update((list) => [...list,message]);
}

export function clearMessage(){
  messages.set([]);
}
