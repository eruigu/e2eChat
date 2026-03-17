<script lang="ts">
  import { onDestroy } from "svelte";
  import { get } from "svelte/store";
  import { chatSocket } from "$lib/chat/socket";
  import { messages, status, username, onlineCount } from "$lib/chat/store";
  import type { IncomingMessage } from "$lib/chat/types";

  let name = $state("");
  let text = $state("");
  let hasJoined = $state(false);

  function joinChat() {
    const trimmed = name.trim();
    if (!trimmed) return;

    username.set(trimmed);
    chatSocket.connect(trimmed);
    hasJoined = true;
  }

  function sendMessage() {
    const user = get(username);
    const trimmed = text.trim();
    if (!user || !trimmed) return;

    const msg: IncomingMessage = {
      user,
      text: trimmed,
      time: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false
      })
    };

    const sent = chatSocket.send(msg);
    if (sent) text = "";
  }

  onDestroy(() => {
    chatSocket.disconnect();
  });
</script>

{#if !hasJoined}
  <div class="page">
    <div class="welcome-card">
      <h1>Welcome</h1>
      <p>Secure channel — enter your name to chat</p>

      <input bind:value={name} placeholder="Enter your name" />
      <button onclick={joinChat}>Enter Chat</button>
    </div>
  </div>
{:else}
  <div class="page">
    <div class="chat-shell">
      <div class="chat-header">
        <h2>Chat</h2>
        <p>Online: {$onlineCount}</p>
      </div>

      <div class="chat-messages">
        {#each $messages as msg}
          <div class="message">
            <span class="time">[{msg.time}]</span>
            <strong>{msg.user}:</strong>
            <span>{msg.text}</span>
          </div>
        {/each}
      </div>

      <div class="chat-composer">
        <input
          bind:value={text}
          placeholder="Type a message"
          onkeydown={(e) => e.key === "Enter" && sendMessage()}
        />
        <button onclick={sendMessage} disabled={$status !== "connected"}>
          Send
        </button>
      </div>
    </div>
  </div>
{/if}
