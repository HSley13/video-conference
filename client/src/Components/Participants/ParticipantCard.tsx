import { User } from "../../Types/types";
import { Pin, PinOff, Mic, MicOff, Video, VideoOff } from "lucide-react";
import { Button, Card, Row, Col } from "react-bootstrap";

type ParticipantCardProps = {
  user: User;
  onPin: (id: string) => void;
  onVideoToggle: (id: string) => void;
  onAudioToggle: (id: string) => void;
};
export const ParticipantCard = ({
  user,
  onPin,
  onVideoToggle,
  onAudioToggle,
}: ParticipantCardProps) => {
  return (
    <Card
      className={`mb-2 border-2 ${user.isPinned ? "border-primary" : "border-light"} rounded-pill`}
    >
      <Card.Body className="p-2">
        <Row className="align-items-center g-2">
          <Col xs="auto" className="flex-grow-1">
            <div className="d-flex align-items-center gap-2">
              <img
                src={user.imgUrl}
                alt={user.name}
                className="rounded-circle"
                width={40}
                height={40}
              />
              <span className="font-medium text-truncate">{user.name}</span>
            </div>
          </Col>

          <Col xs="auto">
            <div className="d-flex gap-2">
              <Button
                variant="link"
                size="sm"
                onClick={() => onPin(user.id.toString())}
                aria-label={user.isPinned ? "Unpin" : "Pin"}
                className="text-decoration-none p-0"
              >
                {user.isPinned ? (
                  <Pin className="text-primary" size={20} />
                ) : (
                  <PinOff className="text-danger" size={20} />
                )}
              </Button>

              <Button
                variant="link"
                size="sm"
                onClick={() => onVideoToggle(user.id.toString())}
                aria-label={user.isVideoOn ? "Turn off video" : "Turn on video"}
                className="text-decoration-none p-0"
              >
                {user.isVideoOn ? (
                  <Video className="text-primary" size={20} />
                ) : (
                  <VideoOff className="text-danger" size={20} />
                )}
              </Button>

              <Button
                variant="link"
                size="sm"
                onClick={() => onAudioToggle(user.id.toString())}
                aria-label={user.isAudioOn ? "Mute" : "Unmute"}
                className="text-decoration-none p-0"
              >
                {user.isAudioOn ? (
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
