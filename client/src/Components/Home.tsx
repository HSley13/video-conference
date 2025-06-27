import { VideoWindow } from "./VideoWindow/VideoWindow";
import { useState } from "react";
import { Container, Row, Col, Form, Button } from "react-bootstrap";
import { useNavigate } from "react-router-dom";

export const Home = () => {
  const [newRoom, setNewRoom] = useState("");
  const [joinRoom, setJoinRoom] = useState("");

  const handleCreateRoom = () => {
    setNewRoom(newRoom);
  };

  const handleJoinRoom = () => {
    setJoinRoom(joinRoom);
  };

  return (
    <Container
      fluid
      className="d-flex align-center justify-content-center vh-100 bg-light"
    >
      <Row style={{ maxWidth: 480 }} className="w-100 gy-4">
        <Col xs={12}>
          {" "}
          <h4> Create a New Conference</h4>{" "}
          <Form.Control
            type="text"
            placeholder="Room Name"
            value={newRoom}
            onChange={(e) => setNewRoom(e.target.value)}
            className="mb-3"
          />
          <Button
            variant="primary"
            className="w-100"
            onClick={handleCreateRoom}
          >
            Create &amp; Start
          </Button>
        </Col>
        <Col xs={12}>
          <h4 className="mt-4"> Join a Conference</h4>
          <Form.Control
            type="text"
            placeholder="Paste Room link"
            value={joinRoom}
            onChange={(e) => setJoinRoom(e.target.value)}
            className="mb-3"
          />
          <Button variant="success" className="w-100" onClick={handleJoinRoom}>
            Join
          </Button>
        </Col>
      </Row>
    </Container>
  );
};
