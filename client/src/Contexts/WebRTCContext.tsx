import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useCallback,
  Dispatch,
  SetStateAction,
  useState,
  ReactNode,
} from "react";
import { User } from "../Types/types";

type RemoteStreamsMap = Record<string, MediaStream>;

type WebRTCContextValue = {
  localStream: MediaStream | null;
  remoteStreams: RemoteStreamsMap;
  users: User[];
  addRemoteStream: (userId: string, stream: MediaStream) => void;
  removeRemoteStream: (userId: string) => void;
  setUsers: Dispatch<SetStateAction<User[]>>;
};

const WebRTCContext = createContext<WebRTCContextValue | undefined>(undefined);

export const WebRTCProvider = ({ children }: { children: ReactNode }) => {
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<RemoteStreamsMap>({});
  const [users, setUsers] = useState<User[]>([]);

  const hasRequestedMedia = useRef(false);

  const addRemoteStream = useCallback((userId: string, stream: MediaStream) => {
    setRemoteStreams((prev) => ({ ...prev, [userId]: stream }));
    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, videoStream: stream } : u)),
    );
  }, []);

  const removeRemoteStream = useCallback((userId: string) => {
    setRemoteStreams((prev) => {
      const { [userId]: removed, ...rest } = prev;
      removed?.getTracks().forEach((t) => t.stop());
      return rest;
    });

    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, videoStream: null } : u)),
    );
  }, []);

  useEffect(() => {
    if (hasRequestedMedia.current) return;
    hasRequestedMedia.current = true;

    const obtainMedia = async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: true,
        });
        setLocalStream(stream);
      } catch (err) {
        console.error("[WebRTC] getUserMedia failed:", err);
      }
    };

    obtainMedia();

    return () => {
      setLocalStream((current) => {
        current?.getTracks().forEach((t) => t.stop());
        return null;
      });

      setRemoteStreams((prev) => {
        Object.values(prev).forEach((s) =>
          s.getTracks().forEach((t) => t.stop()),
        );
        return {};
      });
    };
  }, []);

  const value: WebRTCContextValue = {
    localStream,
    remoteStreams,
    users,
    addRemoteStream,
    removeRemoteStream,
    setUsers,
  };

  return (
    <WebRTCContext.Provider value={value}>{children}</WebRTCContext.Provider>
  );
};

export const useWebRTC = (): WebRTCContextValue => {
  const ctx = useContext(WebRTCContext);
  if (!ctx) {
    throw new Error("useWebRTC must be used inside <WebRTCProvider>");
  }
  return ctx;
};
