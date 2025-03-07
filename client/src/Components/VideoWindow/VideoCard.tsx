import { useEffect, useRef } from "react";
import { Card, Button, Col } from "react-bootstrap";
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
    <Card className="h-100 position-relative overflow-hidden bg-secondary">
      <Card.Body className="p-0 h-100">
        {videoStream ? (
          <video
            ref={videoRef}
            className="h-100 w-100 object-cover"
            autoPlay
            playsInline
          />
        ) : (
          <Card.Img
            variant="top"
            src={imgUrl}
            alt={name}
            className="h-100 w-100 object-cover"
          />
        )}
      </Card.Body>

      <audio ref={audioRef} autoPlay playsInline />

      <Col className="position-absolute bottom-0 start-0 w-100 m-0 px-2 pb-2">
        <span className="badge bg-dark bg-opacity-75 text-white rounded-pill px-3">
          {name}
        </span>
      </Col>

      <Col className="position-absolute top-0 end-0 m-1 gap-2 d-flex flex-row align-items-center justify-content-end ">
        <Button
          variant="dark"
          size="sm"
          onClick={() => onPin(id)}
          className="rounded-circle p-1 bg-opacity-75"
          aria-label={isPinned ? "Unpin" : "Pin"}
        >
          {isPinned ? (
            <PinOff size={16} className="m-1" />
          ) : (
            <Pin size={16} className="m-1" />
          )}
        </Button>

        <Button
          variant="dark"
          size="sm"
          className="rounded-circle p-1 bg-opacity-75"
          disabled={!isAudioOn}
        >
          {!isAudioOn ? (
            <MicOff size={16} className="m-1" />
          ) : (
            <Mic size={16} className="m-1" />
          )}
        </Button>
      </Col>
    </Card>
  );
};
