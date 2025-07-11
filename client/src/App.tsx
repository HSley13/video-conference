import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import { MainWindow } from "./Components/MainWindow";
import { VideoWindow } from "./Components/VideoWindow/VideoWindow";
import { Route, Routes, Navigate } from "react-router-dom";
import { Container } from "react-bootstrap";
import { WebRTCProvider } from "./Contexts/WebRTCContext";
import { Authentification } from "./Components/Authentification/Authentification";

export const App = () => {
  return (
    <Container>
      <WebRTCProvider>
        <Routes>
          <Route path="/" element={<Authentification />} />
          <Route path="/mainwindow" element={<MainWindow />} />
          <Route path="/videoWindow" element={<VideoWindow />} />
          <Route path="/profile" element={<h1>Profile</h1>} />
          <Route path="/settings" element={<h1>Settings</h1>} />
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </WebRTCProvider>
    </Container>
  );
};
