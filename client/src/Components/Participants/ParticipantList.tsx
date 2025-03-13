import { ParticipantCard } from "./ParticipantCard";
import { useWebRTC } from "../../Contexts/WebRTCContext";

export const ParticipantList = () => {
  const { users: Participants, setUsers: setParticipants } = useWebRTC();

  return (
    <div className="flex flex-col h-[320px] w-80 bg-white rounded-lg overflow-hidden">
      <h3 className="text-lg font-semibold m-2">Participants</h3>
      <div className="flex-1 overflow-y-auto rounded-lg p-2 bg-gray-200">
        {Participants.map((participant) => (
          <ParticipantCard
            key={participant.id}
            user={participant}
            onPin={(id) => {
              const updatedParticipants = Participants.map((p) => {
                if (p.id === parseInt(id)) {
                  return { ...p, isPinned: !p.isPinned };
                }
                return p;
              });
              setParticipants(updatedParticipants);
            }}
            onVideoToggle={(id) => {
              const updatedParticipants = Participants.map((p) => {
                if (p.id === parseInt(id)) {
                  return { ...p, isVideoOn: !p.isVideoOn };
                }
                return p;
              });
              setParticipants(updatedParticipants);
            }}
            onAudioToggle={(id) => {
              const updatedParticipants = Participants.map((p) => {
                if (p.id === parseInt(id)) {
                  return { ...p, isAudioOn: !p.isAudioOn };
                }
                return p;
              });
              setParticipants(updatedParticipants);
            }}
          />
        ))}
      </div>
    </div>
  );
};
