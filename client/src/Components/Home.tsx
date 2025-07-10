import { VideoWindow } from "./VideoWindow/VideoWindow";
import { useState } from "react";
import { Container, Row, Col, Form, Button } from "react-bootstrap";
import { useNavigate } from "react-router-dom";
import { createRoom } from "../Services/room";
import { useAsyncFn } from "../Hooks/useAsync";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

export const Home = () => {
  const [roomName, setRoomName] = useState("");
  const [roomLink, setRoomLink] = useState("");
  // const [showVideo, setShowVideo] = useState(false);
  const createRoomFn = useAsyncFn(createRoom);

  const navigate = useNavigate();

  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!roomName.trim()) return;

    const createRoomResponse = await createRoomFn.execute({
      title: roomName,
      description: "This is the Avengers Group",
    });

    toast.success("Room created: " + createRoomResponse.id);
  };

  const handleJoinRoom = () => {
    if (!roomLink.trim()) return;
    navigate(roomLink.trim());
  };

  // if (showVideo) {
  // return <VideoWindow />;
  // }

  return (
    <>
      {/* <Button className="mb-3" onClick={() => setShowVideo(true)}> */}
      {/* VideoConference */}
      {/* </Button> */}

      <Container
        fluid
        className="d-flex align-items-center justify-content-center vh-100 bg-light"
      >
        <Row style={{ maxWidth: 480 }} className="w-100 gy-4">
          <Col xs={12}>
            <h4>Create a New Conference</h4>
            <Form.Control
              type="text"
              placeholder="Room Name"
              value={roomName}
              onChange={(e) => setRoomName(e.target.value)}
              className="mb-3"
            />
            <Button
              variant="primary"
              className="w-100"
              onClick={handleCreateRoom}
            >
              Create&nbsp;&amp;&nbsp;Start
            </Button>
            <ToastContainer position="top-center" />
          </Col>

          <Col xs={12}>
            <h4 className="mt-4">Join a Conference</h4>
            <Form.Control
              type="text"
              placeholder="Paste room link"
              value={roomLink}
              onChange={(e) => setRoomLink(e.target.value)}
              className="mb-3"
            />
            <Button
              variant="success"
              className="w-100"
              onClick={handleJoinRoom}
            >
              Join
            </Button>
          </Col>
        </Row>
      </Container>
    </>
  );
};
