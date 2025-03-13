// Contexts/WebRTCContext.ts
import { createContext, useContext, useEffect, useState } from "react";
import { User } from "../Types/types";

type WebRTCContextType = {
  localStream: MediaStream | null;
  remoteStreams: Record<number, MediaStream>;
  users: User[];
  addRemoteStream: (userId: number, stream: MediaStream) => void;
  removeRemoteStream: (userId: number) => void;
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

export const WebRTCProvider = ({ children }: { children: React.ReactNode }) => {
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<
    Record<number, MediaStream>
  >({});
  const [users, setUsers] = useState<User[]>([
    {
      id: 1,
      name: "John",
      imgUrl: "https://randomuser.me/api/portraits/men/1.jpg",
      isAudioOn: true,
      isPinned: true,
      videoStream: null,
    },
    {
      id: 2,
      name: "Sarah",
      imgUrl: "https://randomuser.me/api/portraits/women/1.jpg",
      isAudioOn: false,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 3,
      name: "Alex",
      imgUrl: "https://randomuser.me/api/portraits/men/2.jpg",
      isAudioOn: true,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 4,
      name: "Maria",
      imgUrl: "https://randomuser.me/api/portraits/women/2.jpg",
      isAudioOn: true,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 5,
      name: "Emily",
      imgUrl: "https://randomuser.me/api/portraits/women/3.jpg",
      isAudioOn: false,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 6,
      name: "Michael",
      imgUrl: "https://randomuser.me/api/portraits/men/3.jpg",
      isAudioOn: false,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 7,
      name: "Olivia",
      imgUrl: "https://randomuser.me/api/portraits/women/4.jpg",
      isAudioOn: true,
      isPinned: false,
      videoStream: null,
    },
    {
      id: 8,
      name: "William",
      imgUrl: "https://randomuser.me/api/portraits/men/4.jpg",
      isAudioOn: true,
      isPinned: false,
      videoStream: null,
    },
  ]);

  const addRemoteStream = (userId: number, stream: MediaStream) => {
    setRemoteStreams((prevStreams) => ({
      ...prevStreams,
      [userId]: stream,
    }));

    setUsers((prevUsers) => {
      const existingUser = prevUsers.find((user) => user.id === userId);
      if (existingUser) {
        return prevUsers.map((user) =>
          user.id === userId ? { ...user, videoStream: stream } : user,
        );
      }
      return [
        ...prevUsers,
        {
          id: userId,
          name: users.length > 0 ? users[0].name : "",
          imgUrl: users.length > 0 ? users[0].imgUrl : "",
          isAudioOn: true,
          isPinned: false,
          isVideoOn: true,
          videoStream: stream,
        },
      ];
    });
  };

  const removeRemoteStream = (userId: number) => {
    setRemoteStreams((prevStreams) => {
      const updatedStreams = { ...prevStreams };
      delete updatedStreams[userId];
      return updatedStreams;
    });

    setUsers((prevUsers) =>
      prevUsers.map((user) =>
        user.id === userId ? { ...user, videoStream: null } : user,
      ),
    );
  };

  useEffect(() => {
    navigator.mediaDevices
      .getUserMedia({
        video: true,
        audio: true,
      })
      .then((stream) => {
        setLocalStream(stream);
      })
      .catch((error) => {
        console.error("Error accessing media devices:", error);
      });

    return () => {
      if (localStream) {
        localStream.getTracks().forEach((track) => track.stop());
      }
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
  const context = useContext(WebRTCContext);
  if (!context) {
    throw new Error("useWebRTC must be used within a WebRTCProvider");
  }
  return context;
};
