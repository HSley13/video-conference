import { useState, useEffect, useRef } from "react";
import { Mic, MicOff, Pin, PinOff } from "lucide-react";

type VideoCardProps = {
  id: number;
  name: string;
  isAudioOn: boolean;
  isPinned: boolean;
  imgUrl: string;
  videoStream: MediaStream | null;
  onPin: (id: number) => void;
};

export const VideoCard = ({
  id,
  name,
  isAudioOn,
  isPinned,
  imgUrl,
  videoStream,
  onPin,
}: VideoCardProps) => {
  const [isMuted, setIsMuted] = useState(true);
  const videoRef = useRef<HTMLVideoElement>(null);
  const audioRef = useRef<HTMLAudioElement>(null);

  useEffect(() => {
    if (videoRef.current && videoStream) {
      videoRef.current.srcObject = videoStream;
      videoRef.current.play().catch(console.error);
    }
  }, [videoStream]);

  useEffect(() => {
    if (audioRef.current && videoStream) {
      const audioTracks = videoStream.getAudioTracks();
      if (audioTracks.length > 0) {
        const audioStream = new MediaStream(audioTracks);
        audioRef.current.srcObject = audioStream;
        audioRef.current.play().catch(console.error);
      }
    }
  }, [videoStream]);

  return (
    <div className="relative group w-full h-full rounded-lg overflow-hidden shadow-lg hover:shadow-xl transition-all duration-200">
      {videoStream ? (
        <video
          ref={videoRef}
          className="w-full h-full object-cover"
          muted={isMuted}
          autoPlay
          playsInline
        />
      ) : (
        <div className="w-full h-full bg-gray-200 flex items-center justify-center">
          <img src={imgUrl} alt={name} className="w-full h-full object-cover" />
        </div>
      )}

      <audio ref={audioRef} muted={isMuted} autoPlay playsInline />

      <div className="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/60 to-transparent">
        <span className="text-white font-medium truncate">{name}</span>
      </div>

      <div className="absolute top-0 right-0 p-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <button
          onClick={() => onPin(id)}
          className="p-2 rounded-full bg-black/50 hover:bg-black/70 transition-colors text-white"
          aria-label={isPinned ? "Unpin" : "Pin"}
        >
          {isPinned ? <PinOff size={20} /> : <Pin size={20} />}
        </button>

        <button
          onClick={() => setIsMuted(!isMuted)}
          className="p-2 rounded-full bg-black/50 hover:bg-black/70 transition-colors text-white"
          aria-label={isMuted ? "Unmute" : "Mute"}
          disabled={!isAudioOn}
        >
          {isMuted || !isAudioOn ? <MicOff size={20} /> : <Mic size={20} />}
        </button>
      </div>
    </div>
  );
};
