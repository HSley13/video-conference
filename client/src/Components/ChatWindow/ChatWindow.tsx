import { useState, useRef, useEffect } from "react";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { useVideoConference } from "../../Hooks/websocket";

export const ChatWindow = () => {
  const userID = 123;
  const userName = "Sley";
  const userPhoto = "https://randomuser.me/api/portraits/men/2.jpg";
  const roomID = "123";
  const [newMessage, setNewMessage] = useState("");
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { chatMessages, sendChatMessage } = useVideoConference(
    roomID,
    userID,
    userName,
    userPhoto,
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
        {chatMessages.map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            isCurrentUser={userID === message.user.id}
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
