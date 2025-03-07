import { Participant } from "../../Types/types";
import { Pin, PinOff, Mic, MicOff, Video, VideoOff } from "lucide-react";
import { Button, Card, Row, Col } from "react-bootstrap";

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
    <Card
      className={`mb-2 border-2 ${participant.isPinned ? "border-primary" : "border-light"} rounded-pill`}
    >
      <Card.Body className="p-2">
        <Row className="align-items-center g-2">
          <Col xs="auto" className="flex-grow-1">
            <div className="d-flex align-items-center gap-2">
              <img
                src={participant.photo}
                alt={participant.name}
                className="rounded-circle"
                width={40}
                height={40}
              />
              <span className="font-medium text-truncate">
                {participant.name}
              </span>
            </div>
          </Col>

          <Col xs="auto">
            <div className="d-flex gap-2">
              <Button
                variant="link"
                size="sm"
                onClick={() => onPin(participant.id.toString())}
                aria-label={participant.isPinned ? "Unpin" : "Pin"}
                className="text-decoration-none p-0"
              >
                {participant.isPinned ? (
                  <Pin className="text-primary" size={20} />
                ) : (
                  <PinOff className="text-danger" size={20} />
                )}
              </Button>

              <Button
                variant="link"
                size="sm"
                onClick={() => onVideoToggle(participant.id.toString())}
                aria-label={
                  participant.videoOn ? "Turn off video" : "Turn on video"
                }
                className="text-decoration-none p-0"
              >
                {participant.videoOn ? (
                  <Video className="text-primary" size={20} />
                ) : (
                  <VideoOff className="text-danger" size={20} />
                )}
              </Button>

              <Button
                variant="link"
                size="sm"
                onClick={() => onAudioToggle(participant.id.toString())}
                aria-label={participant.audioOn ? "Mute" : "Unmute"}
                className="text-decoration-none p-0"
              >
                {participant.audioOn ? (
                  <Mic className="text-primary" size={20} />
                ) : (
                  <MicOff className="text-danger" size={20} />
                )}
              </Button>
            </div>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );
};
