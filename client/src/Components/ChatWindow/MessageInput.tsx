import { Card, Button, Row, Col, Form } from "react-bootstrap";
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
    <Card className="border-gray-200 rounded-bottom-0">
      <Card.Body className="p-2 bg-white">
        {showEmojiPicker && (
          <div className="position-absolute bottom-100 start-0">
            <EmojiPicker onEmojiClick={handleEmojiClick} />
          </div>
        )}

        <Row className="align-items-center g-2">
          <Col xs="auto">
            <Button
              variant="link"
              onClick={() => setShowEmojiPicker(!showEmojiPicker)}
              className="text-secondary p-0"
            >
              <Smile size={20} />
            </Button>
          </Col>

          <Col>
            <Form.Control
              type="text"
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              onKeyPress={(e) => e.key === "Enter" && handleSend(newMessage)}
              placeholder="Type a message..."
              className="rounded-pill border-1 bg-light"
            />
          </Col>

          <Col xs="auto">
            <Button
              variant="primary"
              onClick={() => handleSend(newMessage)}
              className="rounded-circle p-2"
            >
              <SendHorizontal size={20} />
            </Button>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );
};
