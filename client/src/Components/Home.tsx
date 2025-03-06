import { SideBar } from "./SideBar/SideBar";
import { ChatWindow } from "./ChatWindow/ChatWindow";
import { ParticipantList } from "./Participants/ParticipantList";

export const Home = () => {
  return (
    <div className="flex flex-row h-screen relative">
      <div className="fixed top-0 left-0 h-full">
        <SideBar />
      </div>

      <div className="fixed top-0 right-0 p-3">
        <ParticipantList />
      </div>

      <div className="fixed bottom-0 right-0 p-3">
        <ChatWindow />
      </div>
    </div>
  );
};
