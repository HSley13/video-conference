type Message = {
  id: number;
  text: string;
  time: string;
  user: {
    id: number;
    name: string;
    photo: string;
  };
};

type MessageBubbleProps = {
  message: Message;
  isCurrentUser: boolean;
};
export const MessageBubble = ({
  message,
  isCurrentUser,
}: MessageBubbleProps) => {
  return (
    <div
      className={`d-flex mb-3 ${isCurrentUser ? "justify-content-end" : "justify-content-start"}`}
    >
      <div
        className={`d-flex align-items-center ${isCurrentUser ? "flex-row-reverse" : ""}`}
        style={{ maxWidth: "95%" }}
      >
        <img
          src={message.user.photo}
          alt={message.user.name}
          className="rounded-circle flex-shrink-0"
          style={{ width: "40px", height: "40px" }}
        />

        <div
          className={`ms-2 me-2 ${isCurrentUser ? "text-end" : "text-start"}`}
          style={{ minWidth: "100px" }}
        >
          <div className="d-flex justify-content-between align-items-center mb-1">
            <small className="fw-bold text-truncate">{message.user.name}</small>
            <small className="text-muted ms-2">{message.time}</small>
          </div>
          <div
            className={`p-3 rounded-4 ${
              isCurrentUser ? "bg-blue-100" : "bg-white"
            }`}
            style={{
              wordWrap: "break-word",
              overflowWrap: "break-word",
              whiteSpace: "pre-wrap",
            }}
          >
            {message.text}
          </div>
        </div>
      </div>
    </div>
  );
};
