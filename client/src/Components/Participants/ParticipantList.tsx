import { useState } from "react";
import { ParticipantCard } from "./ParticipantCard";
import { Participant } from "../../Types/types";

export const ParticipantList = () => {
  const [Participants, setParticipants] = useState<Participant[]>([
    {
      id: 1,
      name: "John",
      photo: "https://randomuser.me/api/portraits/men/1.jpg",
      isPinned: true,
      videoOn: true,
      audioOn: true,
    },
    {
      id: 2,
      name: "Sarah",
      photo: "https://randomuser.me/api/portraits/women/1.jpg",
      isPinned: false,
      videoOn: false,
      audioOn: false,
    },
    {
      id: 3,
      name: "Alex",
      photo: "https://randomuser.me/api/portraits/men/2.jpg",
      isPinned: false,
      videoOn: false,
      audioOn: false,
    },
    {
      id: 4,
      name: "Maria",
      photo: "https://randomuser.me/api/portraits/women/2.jpg",
      isPinned: false,
      videoOn: false,
      audioOn: false,
    },
    {
      id: 5,
      name: "John",
      photo: "https://randomuser.me/api/portraits/men/1.jpg",
      isPinned: false,
      videoOn: false,
      audioOn: false,
    },
  ]);

  return (
    <div className="flex flex-col h-[300px] w-95 bg-white rounded-lg overflow-hidden">
      <h3 className="text-lg font-semibold mb-2">Participants</h3>
      <div className="flex-1 overflow-y-auto rounded-lg p-2 bg-gray-200">
        {Participants.map((participant) => (
          <ParticipantCard
            participant={participant}
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
                  return { ...p, videoOn: !p.videoOn };
                }
                return p;
              });
              setParticipants(updatedParticipants);
            }}
            onAudioToggle={(id) => {
              const updatedParticipants = Participants.map((p) => {
                if (p.id === parseInt(id)) {
                  return { ...p, audioOn: !p.audioOn };
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
