import { SideBar } from "../SideBar/SideBar";
import { ChatWindow } from "../ChatWindow/ChatWindow";
import { ParticipantList } from "../Participants/ParticipantList";
import { useState } from "react";
import { ChevronRight, ChevronLeft, Users, MessageCircle } from "lucide-react";
import { VideoList } from "./VideoList";
import { Row } from "react-bootstrap";

export const MainWindow = () => {
  const [showParticipants, setShowParticipants] = useState(false);
  const [showChat, setShowChat] = useState(false);

  return (
    <Row className="m-0 p-0 gap-0 h-screen w-screen">
      <div className="flex flex-row gap-0  w-full h-full">
        <div className="fixed top-0 left-0 h-full">
          <SideBar />
        </div>

        <div className="flex-grow m-0 p-0 overflow-hidden">
          <VideoList />
        </div>

        <div className="fixed top-0 right-0 h-full z-10">
          <button
            onClick={() => setShowParticipants(!showParticipants)}
            className="fixed top-2 right-4 p-2 bg-indigo-600 text-white rounded-full shadow-lg hover:bg-indigo-700 transition-all duration-300 flex items-center gap-2 group"
          >
            <Users className="w-5 h-5" />
            <span className="text-sm font-medium hidden md:inline">
              Participants
            </span>
            {showParticipants ? (
              <ChevronLeft className="w-4 h-4 transform group-hover:scale-110" />
            ) : (
              <ChevronRight className="w-4 h-4 transform group-hover:scale-110" />
            )}
          </button>

          <div
            className={`fixed top-12 right-4 transform transition-all duration-300 ease-out ${
              showParticipants
                ? "translate-x-0 opacity-100"
                : "translate-x-full opacity-0 pointer-events-none"
            }`}
          >
            <ParticipantList />
          </div>

          <button
            onClick={() => setShowChat(!showChat)}
            className="fixed bottom-2 right-4 p-2 bg-emerald-600 text-white rounded-full shadow-lg hover:bg-emerald-700 transition-all duration-300 flex items-center gap-2 group"
          >
            <MessageCircle className="w-5 h-5" />
            <span className="text-sm font-medium hidden md:inline">Chat</span>
            {showChat ? (
              <ChevronLeft className="w-4 h-4 transform group-hover:scale-110" />
            ) : (
              <ChevronRight className="w-4 h-4 transform group-hover:scale-110" />
            )}
          </button>

          <div
            className={`fixed bottom-12 right-4 transform transition-all duration-300 ease-out ${
              showChat
                ? "translate-x-0 opacity-100"
                : "translate-x-full opacity-0 pointer-events-none"
            }`}
          >
            <ChatWindow />
          </div>
        </div>
      </div>
    </Row>
  );
};
