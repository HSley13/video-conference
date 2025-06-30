import { useState, useRef, useEffect } from "react";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { useVideoConference } from "../../Hooks/websocket";
import { useWebRTC } from "../../Contexts/WebRTCContext";

export const ChatWindow = () => {
  const [draft, setDraft] = useState("");
  const [showEmojiPicker, setPickerOpen] = useState(false);
  const endRef = useRef<HTMLDivElement>(null);
  const { userInfo } = useWebRTC();
  const { messages: chatMessages, sendChatMessage } = useVideoConference();

  useEffect(() => {
    requestAnimationFrame(() =>
      endRef.current?.scrollIntoView({ behavior: "smooth" }),
    );
  }, [chatMessages]);

  const handleSend = (text: string) => {
    const msg = text.trim();
    if (!msg) return;
    sendChatMessage(msg);
    setDraft("");
    setPickerOpen(false);
  };

  return (
    <div className="flex flex-col h-[400px] w-80 bg-white rounded-lg overflow-hidden">
      <h3 className="text-lg font-semibold m-2">Chat</h3>

      <div className="flex-1 overflow-y-auto rounded-t-lg p-2 bg-gray-200">
        {chatMessages.map((m) => (
          <MessageBubble
            key={m.id}
            message={m}
            isCurrentUser={m.user.id === userInfo.id}
          />
        ))}
        <div ref={endRef} />
      </div>

      <MessageInput
        newMessage={draft}
        setNewMessage={setDraft}
        handleSend={handleSend}
        showEmojiPicker={showEmojiPicker}
        setShowEmojiPicker={setPickerOpen}
        handleEmojiClick={({ emoji }) => setDraft((p) => p + emoji)}
      />
    </div>
  );
};
