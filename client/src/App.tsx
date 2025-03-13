import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import { MainWindow } from "./Components/VideoWindow/MainWindow";
import { Home } from "./Components/Home";
import { Route, Routes, Navigate } from "react-router-dom";
import { Container } from "react-bootstrap";
import { WebRTCProvider } from "./Contexts/WebRTCContext";

export const App = () => {
  return (
    <Container>
      <WebRTCProvider>
        <Routes>
          <Route path="/" element={<Home />} />
          {/* <Route path="/" element={<MainWindow />} /> */}
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </WebRTCProvider>
    </Container>
  );
};
