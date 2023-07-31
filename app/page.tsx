"use client";

import { io as ClientIO } from "socket.io-client";
import React, { useState, useEffect, useRef, useCallback } from "react";

interface IChatMessage {
  userName: string;
  message: string;
}

const titlecase = (str: string) => {
  return str.toLowerCase().split(' ').map((word: string) => {
    return (word.charAt(0).toUpperCase() + word.slice(1));
  }).join(' ');
}

// component
const Index: React.FC = () => {
  const inputRef = useRef(null);

  // connected flag
  const [connected, setConnected] = useState<boolean>(false);

  // user count
  const [userCount, setUserCount] = useState<number>(0);

  // init chat and message
  const [userName, setUserName] = useState<string>("");
  const [messageInput, setMessageInput] = useState<string>("");
  const [userNameInput, setUserNameInput] = useState<string>("");
  const [chatMessages, setChatMessages] = useState<IChatMessage[]>([]);

  // dispatch message to other users
  const sendApiSocketChat = useCallback(async (
    chatMessage: IChatMessage
  ): Promise<Response> => {
    return await fetch("/api/chat", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(chatMessage),
    });
  },
    []
  );

  const sendMessage = async () => {
    if (messageInput) {
      const chatMessage: IChatMessage = {
        userName,
        message: messageInput,
      };
      const resp = await sendApiSocketChat(chatMessage);
      if (resp.ok) setMessageInput("");
    }

    (inputRef?.current as any).focus();
  };

  const sendEnterRoomMessage = useCallback(async () => {
    const chatMessage: IChatMessage = {
      userName: "Moderator",
      message: `${userName} has entered the chat`,
    };

  const resp = await sendApiSocketChat(chatMessage);
  if (resp.ok) {
    // Fetch updated user count after a new user joins
    const userCountResp = await fetch("/api/userCount");
    if (userCountResp.ok) {
      const { count } = await userCountResp.json();
      setUserCount(count);
    }
  } else {
    setTimeout(() => {
      sendEnterRoomMessage();
    }, 500);
  }
}, [userName, sendApiSocketChat]);

  useEffect((): any => {
    if (userName) {
      sendEnterRoomMessage();
    }
  }, [userName, sendEnterRoomMessage]);

  useEffect((): any => {
    const socket = new (ClientIO as any)(process.env.NEXT_PUBLIC_SITE_URL, {
      path: "/api/socket",
      addTrailingSlash: false,
    });

    // log socket connection
    socket.on("connect", () => {
      console.log("SOCKET CONNECTED!", socket.id);
      setConnected(true);
    });

    // update chat on new message dispatched
    socket.on("message", (message: IChatMessage) => {
      chatMessages.push(message);
      setChatMessages([...chatMessages]);
    });

    // update user count on new user connection
    socket.on("userConnect", (count: number) => {
      setUserCount(count);
    });

    // update user count on user disconnection
    socket.on("userDisconnect", (count: number) => {
      setUserCount(count);
    });
    // socket disconnect onUnmount if exists
    if (socket) return () => socket.disconnect();
  }, [chatMessages]);

  if (!connected) {
    return (
      <div className="flex items-center p-4 mx-auto min-h-screen justify-center bg-black">
        <div className="gap-4 flex flex-col items-center justify-center w-full h-full">
          <div className="flex items-center justify-center space-x-2 animate-bounce">
            <div className="w-6 h-6 bg-blue-400 rounded-full"></div>
            <div className="w-6 h-6 bg-green-400 rounded-full"></div>
            <div className="w-6 h-6 bg-red-400 rounded-full"></div>
          </div>
        </div>
      </div>
    );
  }

  if (!userName) {
    return (
      <div className="flex items-center p-4 mx-auto min-h-screen justify-center bg-black">
        <div className="gap-4 flex flex-col items-center justify-center w-full h-full">
          <h3 className="font-bold text-white text-xl">
            Please enter your username
          </h3>
          <div className="bg-black p-4 h-20">
            <div className="flex flex-row flex-1 h-full divide-gray-800 divide-x">
              <div className="pr-2 flex-1">
                <input
                  type="text"
                  value={titlecase(userNameInput)}
                  placeholder={connected ? "Your Name" : "Connecting..."}
                  className="w-full h-full rounded shadow border 
                  border-black-400 px-2 text-black"
                  disabled={!connected}
                  onChange={(e) => setUserNameInput(e.target.value)}
                  onKeyUp={(e) => {
                    if (e.key === "Enter") {
                      setUserName(userNameInput);
                    }
                  }}
                />
              </div>
              <div className="flex flex-col justify-center items-stretch pl-2">
                <button
                  className="bg-gray-600 rounded shadow 
                  text-sm text-white h-full px-2"
                  onClick={() => {
                    setUserName(userNameInput);
                  }}
                  disabled={!connected}
                >
                  Enter
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col w-full h-screen">
      <div className="py-4 text-white bg-black sticky top-0">
        <h1 className="text-center text-2xl font-semibold">
          Real-Time Chat App
        </h1>
        <h2 className="mt-2 text-center">Powered by Next.js, Socket.io, and libsignal</h2>
      </div>
      <div className="flex flex-col flex-1 bg-gray-600">
        <div className="flex items-center justify-center px-4 text-white">
          {connected ? `Users Online: ${userCount}` : "Connecting..."}
        </div>       
        <div className="flex-1 p-4 font-mono">
          {chatMessages.length ? (
            chatMessages.map((chatMessage, i) => (
              <div key={"msg_" + i} className="mt-1 text-white">
                <span
                  className={
                    chatMessage.userName === userName
                      ? `text-red-500`
                      : `text-green-400`
                  }
                >
                  {chatMessage.userName === userName
                    ? "[Me]"
                    : `[${chatMessage.userName}]`}
                </span>
                : {chatMessage.message}
              </div>
            ))
          ) : (
            <div className="text-sm text-center text-gray-400 py-6">
              No chat messages
            </div>
          )}
        </div>
        <div className="bg-black p-4 h-20 sticky bottom-0">
          <div className="flex flex-row flex-1 h-full divide-gray-800 divide-x">
            <div className="pr-2 flex-1">
              <input
                ref={inputRef}
                type="text"
                value={messageInput}
                placeholder={connected ? "Type a message..." : "Connecting..."}
                className="w-full h-full rounded shadow border 
                px-2 border-gray-400 text-black"
                disabled={!connected}
                onChange={(e) => {
                  setMessageInput(e.target.value);
                }}
                onKeyUp={(e) => {
                  if (e.key === "Enter") {
                    sendMessage();
                  }
                }}
              />
            </div>
            <div className="flex flex-col justify-center items-stretch pl-2">
              <button
                className="bg-blue-500 rounded shadow text-sm 
                text-white h-full px-2"
                onClick={() => {
                  sendMessage();
                }}
                disabled={!connected}
              >
                SEND
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Index;
