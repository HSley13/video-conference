import { useState, useRef, useEffect } from "react";
import { Container, Row, Col } from "react-bootstrap";
import { MessageBubble } from "./MessageBubble";
import { MessageInput } from "./MessageInput";
import { useVideoConference } from "../../Hooks/websocket";
import { useWebRTC } from "../../Contexts/WebRTCContext";

export const ChatWindow = () => {
  const [draft, setDraft] = useState("");
  const [pickerOpen, setPickerOpen] = useState(false);
  const endRef = useRef<HTMLDivElement | null>(null);

  const { messages, sendChatMessage } = useVideoConference();

  const { userInfo } = useWebRTC();

  useEffect(() => {
    requestAnimationFrame(() =>
      endRef.current?.scrollIntoView({ behavior: "smooth" }),
    );
  }, [messages]);

  const handleSend = (txt: string) => {
    const value = txt.trim();
    if (!value) return;
    sendChatMessage(value);
    setDraft("");
    setPickerOpen(false);
  };

  return (
    <Container
      className="p-0 rounded shadow"
      style={{ width: 320, height: 520 }}
    >
      <Row className="bg-light border-bottom m-0 py-2">
        <Col>
          <strong>Chat</strong>
        </Col>
      </Row>

      <Row
        className="flex-grow-1 overflow-auto m-0 px-3 py-2"
        style={{ background: "#f3f4f6" }}
      >
        <Col xs={12}>
          {messages.map((m) => (
            <MessageBubble
              key={m.id}
              message={m}
              isCurrentUser={m.user.id === userInfo.id}
            />
          ))}
          <div ref={endRef} />
        </Col>
      </Row>

      <Row className="border-top m-0 p-2">
        <Col xs={12}>
          <MessageInput
            newMessage={draft}
            setNewMessage={setDraft}
            handleSend={handleSend}
            showEmojiPicker={pickerOpen}
            setShowEmojiPicker={setPickerOpen}
            handleEmojiClick={({ emoji }) => setDraft((p) => p + emoji)}
          />
        </Col>
      </Row>
    </Container>
  );
};
