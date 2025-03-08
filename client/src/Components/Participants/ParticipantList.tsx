import { useState } from "react";
import { ParticipantCard } from "./ParticipantCard";
import { User } from "../../Types/types";

export const ParticipantList = () => {
  const [Participants, setParticipants] = useState<User[]>([
    {
      id: 1,
      name: "John",
      imgUrl: "https://randomuser.me/api/portraits/men/1.jpg",
      isAudioOn: true,
      isPinned: true,
      isVideoOn: true,
    },
    {
      id: 2,
      name: "Sarah",
      imgUrl: "https://randomuser.me/api/portraits/women/1.jpg",
      isAudioOn: false,
      isPinned: false,
      isVideoOn: false,
    },
    {
      id: 3,
      name: "Alex",
      imgUrl: "https://randomuser.me/api/portraits/men/2.jpg",
      isAudioOn: false,
      isPinned: false,
      isVideoOn: false,
    },
    {
      id: 4,
      name: "Maria",
      imgUrl: "https://randomuser.me/api/portraits/women/2.jpg",
      isAudioOn: true,
      isPinned: false,
      isVideoOn: true,
    },
  ]);

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
