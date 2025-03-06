import EmojiPicker from "emoji-picker-react";
import { Smile, SendHorizontal } from "lucide-react";

type MessageInputProps = {
  newMessage: string;
  setNewMessage: React.Dispatch<React.SetStateAction<string>>;
  handleSend: (messageText: string) => void;
  showEmojiPicker: boolean;
  setShowEmojiPicker: React.Dispatch<React.SetStateAction<boolean>>;
  handleEmojiClick: (emojiObject: { emoji: string }) => void;
};
export const MessageInput = ({
  newMessage,
  setNewMessage,
  handleSend,
  showEmojiPicker,
  setShowEmojiPicker,
  handleEmojiClick,
}: MessageInputProps) => {
  return (
    <div className="border-3 border-gray-200 p-2 bg-white rounded-b-lg">
      {showEmojiPicker && (
        <div className="absolute bottom-20 left-0">
          <EmojiPicker onEmojiClick={handleEmojiClick} />
        </div>
      )}

      <div className="flex items-center gap-2">
        <button
          onClick={() => setShowEmojiPicker(!showEmojiPicker)}
          className="text-gray-500 hover:text-gray-700"
        >
          <Smile size={20} />
        </button>

        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          onKeyPress={(e) => e.key === "Enter" && handleSend(newMessage)}
          placeholder="Type a message..."
          className="flex-1 border rounded-full py-2 px-4 focus:outline-none focus:border-blue-500"
        />

        <button
          onClick={() => handleSend(newMessage)}
          className="bg-blue-500 text-white p-2 rounded-full hover:bg-blue-600"
        >
          <SendHorizontal size={20} />
        </button>
      </div>
    </div>
  );
};
