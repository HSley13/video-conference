import { Message } from "../../Types/types";
import { Card, Row, Col } from "react-bootstrap";

type MessageBubbleProps = {
  message: Message;
  isCurrentUser: boolean;
};
export const MessageBubble = ({
  message,
  isCurrentUser,
}: MessageBubbleProps) => {
  return (
    <Row
      className={`mb-3 ${isCurrentUser ? "justify-content-end" : "justify-content-start"}`}
    >
      <Col xs={11} md={10} lg={8}>
        <Row className="align-items-center g-2">
          {!isCurrentUser && (
            <Col xs="auto" className="align-self-center">
              <img
                src={message.user.photo}
                alt={message.user.name}
                className="rounded-circle"
                style={{ width: "40px", height: "40px" }}
              />
            </Col>
          )}

          <Col className={`${isCurrentUser ? "text-end" : "text-start"}`}>
            <div
              className={`d-flex justify-content-between ${isCurrentUser ? "flex-row-reverse" : ""}`}
            >
              <small className="fw-bold text-truncate flex-grow-1">
                {message.user.name}
              </small>
              <small className="text-muted ms-2">{message.time}</small>
            </div>

            <Card
              className={`p-3 rounded-4 ${isCurrentUser ? "bg-primary bg-opacity-10" : "bg-white"}`}
            >
              <Card.Text
                className="mb-0 text-break"
                style={{ whiteSpace: "pre-wrap" }}
              >
                {message.text}
              </Card.Text>
            </Card>
          </Col>

          {isCurrentUser && (
            <Col xs="auto" className="align-self-center">
              <img
                src={message.user.photo}
                alt={message.user.name}
                className="rounded-circle"
                style={{ width: "40px", height: "40px" }}
              />
            </Col>
          )}
        </Row>
      </Col>
    </Row>
  );
};
