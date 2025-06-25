import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { User } from "../Types/types";

type WebRTCContextType = {
  localStream: MediaStream | null;
  remoteStreams: Record<string, MediaStream>;
  users: User[];
  addRemoteStream: (userId: string, stream: MediaStream) => void;
  removeRemoteStream: (userId: string) => void;
  setUsers: (users: User[]) => void;
};

const WebRTCContext = createContext<WebRTCContextType>({
  localStream: null,
  remoteStreams: {},
  users: [],
  addRemoteStream: () => {},
  removeRemoteStream: () => {},
  setUsers: () => {},
});

export const WebRTCProvider = ({ children }: { children: ReactNode }) => {
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<
    Record<string, MediaStream>
  >({});
  const [users, setUsers] = useState<User[]>([]);

  const addRemoteStream = (userId: string, stream: MediaStream) => {
    setRemoteStreams((p) => ({ ...p, [userId]: stream }));
    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, videoStream: stream } : u)),
    );
  };

  const removeRemoteStream = (userId: string) => {
    setRemoteStreams((p) => {
      const copy = { ...p };
      delete copy[userId];
      return copy;
    });
    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, videoStream: null } : u)),
    );
  };

  useEffect(() => {
    navigator.mediaDevices
      .getUserMedia({ video: true, audio: true })
      .then(setLocalStream)
      .catch((err) => console.error("getUserMedia:", err));

    return () => {
      localStream?.getTracks().forEach((t) => t.stop());
    };
  }, []);

  return (
    <WebRTCContext.Provider
      value={{
        localStream,
        remoteStreams,
        addRemoteStream,
        removeRemoteStream,
        users,
        setUsers,
      }}
    >
      {children}
    </WebRTCContext.Provider>
  );
};

export const useWebRTC = () => {
  const ctx = useContext(WebRTCContext);
  if (!ctx) throw new Error("useWebRTC must be inside WebRTCProvider");
  return ctx;
};
