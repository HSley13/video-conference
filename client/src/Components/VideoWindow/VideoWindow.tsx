import { ChatWindow } from "../ChatWindow/ChatWindow";
import { ParticipantList } from "../Participants/ParticipantList";
import { useState } from "react";
import { ChevronRight, ChevronLeft, Users, MessageCircle } from "lucide-react";
import { VideoList } from "./VideoList";
import { Row, Col, Button } from "react-bootstrap";

export const VideoWindow = () => {
  const [showParticipants, setShowParticipants] = useState(false);
  const [showChat, setShowChat] = useState(false);

  return (
    <Col className="h-screen w-screen">
      <Row
        className="flex-grow transition-all duration-300 ease-out"
        style={{
          marginLeft: "0",
          marginRight:
            showParticipants || showChat ? `clamp(450px, 35vw, 1120px)` : "0",
        }}
      >
        <VideoList />
      </Row>

      <Button
        onClick={() => setShowParticipants(!showParticipants)}
        className="fixed top-2 right-0 p-2 bg-indigo-600 text-white rounded-full shadow-lg hover:bg-indigo-700 transition-all duration-300 flex items-center gap-2 group"
      >
        <div className="flex items-center gap-2">
          <Users className="w-5 h-5" />
          <div className="h-5 flex items-center">
            {showParticipants ? (
              <ChevronLeft className="w-4 h-4 transform group-hover:scale-110" />
            ) : (
              <ChevronRight className="w-4 h-4 transform group-hover:scale-110" />
            )}
          </div>
        </div>
      </Button>

      <Button
        onClick={() => setShowChat(!showChat)}
        className="fixed bottom-2 right-0 p-2 bg-emerald-600 text-white rounded-full shadow-lg hover:bg-emerald-700 transition-all duration-300 flex items-center gap-2 group"
      >
        <div className="flex items-center gap-2">
          <MessageCircle className="w-5 h-5" />
          <div className="h-5 flex items-center">
            {showChat ? (
              <ChevronLeft className="w-4 h-4 transform group-hover:scale-110" />
            ) : (
              <ChevronRight className="w-4 h-4 transform group-hover:scale-110" />
            )}
          </div>
        </div>
      </Button>

      <Row className="fixed right-0 bottom-0 w-85">
        <div
          className={`fixed top-12 transform transition-all duration-300 ease-out ${
            showParticipants
              ? "translate-x-0 opacity-100"
              : "translate-x-full opacity-0 pointer-events-none"
          }`}
        >
          <ParticipantList />
        </div>

        <div
          className={`fixed bottom-12 transform transition-all duration-300 ease-out ${
            showChat
              ? "translate-x-0 opacity-100"
              : "translate-x-full opacity-0 pointer-events-none"
          }`}
        >
          <ChatWindow />
        </div>
      </Row>
    </Col>
  );
};
