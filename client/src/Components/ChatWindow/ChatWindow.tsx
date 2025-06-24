import { useState, useRef, useEffect } from "react";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { useVideoConference } from "../../Hooks/websocket";

const accessToken =
  localStorage.getItem("access_token") || import.meta.env.VITE_ACCESS_TOKEN;

const userIDEnv = import.meta.env.VITE_USER_ID || undefined;
const userName = import.meta.env.VITE_USER_NAME || "anonymous";
const userPhoto =
  import.meta.env.VITE_USER_PHOTO || "https://via.placeholder.com/40";

const roomID = "550e8400-e29b-41d4-a716-446655440000";

export const ChatWindow = () => {
  const [newMessage, setNewMessage] = useState("");
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { chatMessages, sendChatMessage } = useVideoConference(
    roomID,
    userIDEnv,
    userName,
    userPhoto,
    accessToken,
  );

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [chatMessages]);

  const handleSend = (messageText: string) => {
    if (messageText.trim()) {
      sendChatMessage(messageText);
      setNewMessage("");
      setShowEmojiPicker(false);
    }
  };

  const handleEmojiClick = (emojiObject: { emoji: string }) => {
    setNewMessage((prev) => prev + emojiObject.emoji);
  };

  return (
    <div className="flex flex-col h-[520px] w-80 bg-white rounded-lg overflow-hidden">
      <h3 className="text-lg font-semibold m-2">Chat</h3>

      <div className="flex-1 overflow-y-auto rounded-t-lg p-2 bg-gray-200">
        {chatMessages.map((m) => (
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
