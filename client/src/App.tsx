import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import { MainWindow } from "./Components/MainWindow";
import { Route, Routes, Navigate } from "react-router-dom";
import { Container } from "react-bootstrap";
import { WebRTCProvider } from "./Contexts/WebRTCContext";

export const App = () => {
  return (
    <Container>
      <WebRTCProvider>
        <Routes>
          <Route path="/" element={<MainWindow />} />
          <Route path="/profile" element={<h1>Profile</h1>} />
          <Route path="/settings" element={<h1>Settings</h1>} />
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </WebRTCProvider>
    </Container>
  );
};
