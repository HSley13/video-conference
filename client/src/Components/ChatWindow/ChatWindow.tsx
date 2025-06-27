import { useState, useRef, useEffect } from "react";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { useVideoConference } from "../../Hooks/websocket";

const params = new URLSearchParams(window.location.search);

const envToken = import.meta.env.VITE_ACCESS_TOKEN as string | undefined;
const accessToken = envToken?.trim()
  ? envToken
  : (localStorage.getItem("access_token") ?? "");

const userIDEnv =
  params.get("uid") || import.meta.env.VITE_USER_ID || undefined;
const userName =
  params.get("name") || import.meta.env.VITE_USER_NAME || "anonymous";
const userPhoto =
  params.get("photo") ||
  import.meta.env.VITE_USER_PHOTO ||
  "https://via.placeholder.com/40";

const roomID = params.get("room") || "550e8400-e29b-41d4-a716-446655440000";

export const ChatWindow = () => {
  const [newMessage, setNewMessage] = useState("");
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { messages, sendChatMessage } = useVideoConference(
    roomID,
    userIDEnv,
    userName,
    userPhoto,
    accessToken ?? "",
  );

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (txt: string) => {
    if (txt.trim()) {
      sendChatMessage(txt);
      setNewMessage("");
      setShowEmojiPicker(false);
    }
  };

  const handleEmojiClick = (e: { emoji: string }) =>
    setNewMessage((p) => p + e.emoji);

  return (
    <div className="flex flex-col h-[520px] w-80 bg-white rounded-lg overflow-hidden">
      <h3 className="text-lg fjnt-semibold m-2">Chat</h3>

      <div className="flex-1 overflow-y-auto rounded-t-lg p-2 bg-gray-200">
        {messages.map((m) => (
          <MessageBubble
            key={m.id}
            message={m}
            isCurrentUser={m.user.id === (userIDEnv ?? "")}
          />
        ))}
        <div ref={messagesEndRef} />
      </div>

      <MessageInput
        newMessage={newMessage}
        setNewMessage={setNewMessage}
        handleSend={handleSend}
        showEmojiPicker={showEmojiPicker}
        setShowEmojiPicker={setShowEmojiPicker}
        handleEmojiClick={handleEmojiClick}
      />
    </div>
  );
};
