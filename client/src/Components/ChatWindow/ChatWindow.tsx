import { useState, useRef, useEffect } from "react";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { Message } from "../../Types/types";
import { FormatTime } from "../../Utils/utils";

export const ChatWindow = () => {
  const userID = 123;
  const [newMessage, setNewMessage] = useState("");
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const [messages, setMessages] = useState<Message[]>([
    {
      id: 1,
      text: "Hey, how are you?",
      time: FormatTime(new Date()),
      user: {
        id: 1,
        name: "John",
        photo: "https://randomuser.me/api/portraits/men/1.jpg",
      },
    },
    {
      id: 2,
      text: "I'm good thanks! 😊",
      time: FormatTime(new Date()),
      user: {
        id: 2,
        name: "Sarah",
        photo: "https://randomuser.me/api/portraits/women/1.jpg",
      },
    },
  ]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = (messageText: string) => {
    if (messageText.trim()) {
      const newMsg: Message = {
        id: messages.length + 1,
        text: messageText,
        time: FormatTime(new Date()),
        user: {
          id: userID,
          name: "You",
          photo: "https://randomuser.me/api/portraits/men/2.jpg",
        },
      };
      setMessages([...messages, newMsg]);
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
        {messages.map((message) => (
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
