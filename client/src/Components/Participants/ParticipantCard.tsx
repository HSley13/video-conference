import { Participant } from "../../Types/types";
import { Pin, PinOff, Mic, MicOff, Video, VideoOff } from "lucide-react";

type ParticipantCardProps = {
  participant: Participant;
  onPin: (id: string) => void;
  onVideoToggle: (id: string) => void;
  onAudioToggle: (id: string) => void;
};
export const ParticipantCard = ({
  participant,
  onPin,
  onVideoToggle,
  onAudioToggle,
}: ParticipantCardProps) => {
  return (
    <div
      className={`flex items-center justify-between p-2 bg-white rounded-full mb-2 transition-colors ${
        participant.isPinned ? "border-2 border-blue-200" : ""
      }`}
    >
      <div className="flex items-center gap-2 flex-1 min-w-0">
        <img
          src={participant.photo}
          alt={participant.name}
          className="w-10 h-10 rounded-full object-cover"
        />
        <span className="font-medium truncate">{participant.name}</span>
      </div>

      <div className="flex items-center gap-3 ml-2">
        <button
          onClick={() => onPin(participant.id.toString())}
          className="text-gray-500 hover:text-blue-500 transition-colors"
          aria-label={participant.isPinned ? "Unpin" : "Pin"}
        >
          {participant.isPinned ? (
            <Pin className="w-5 h-5" />
          ) : (
            <PinOff className="w-5 h-5 text-red-500" />
          )}
        </button>

        <button
          onClick={() => onVideoToggle(participant.id.toString())}
          className="text-gray-500 hover:text-blue-500 transition-colors"
          aria-label={participant.videoOn ? "Turn off video" : "Turn on video"}
        >
          {participant.videoOn ? (
            <Video className="w-5 h-5" />
          ) : (
            <VideoOff className="w-5 h-5 text-red-500" />
          )}
        </button>

        <button
          onClick={() => onAudioToggle(participant.id.toString())}
          className="text-gray-500 hover:text-blue-500 transition-colors"
          aria-label={participant.audioOn ? "Mute" : "Unmute"}
        >
          {participant.audioOn ? (
            <Mic className="w-5 h-5" />
          ) : (
            <MicOff className="w-5 h-5 text-red-500" />
          )}
        </button>
      </div>
    </div>
  );
};
