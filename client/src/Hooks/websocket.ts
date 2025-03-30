import { useEffect, useRef, useState } from "react";
import { useWebRTC } from "../Contexts/WebRTCContext";
import { Message } from "../Types/types";
import { FormatTime } from "../Utils/utils";
import { v4 as uuidv4 } from "uuid";

export const useVideoConference = (
  roomID: string,
  userID: string,
  userName: string,
  userPhoto: string,
) => {
  const { localStream, addRemoteStream, removeRemoteStream, setUsers } =
    useWebRTC();
  const wsRef = useRef<WebSocket | null>(null);
  const peerConnections = useRef<Record<string, RTCPeerConnection>>({});

  const [chatMessages, setChatMessages] = useState<Message[]>([
    {
      id: uuidv4(),
      text: "Hey, how are you?",
      time: FormatTime(new Date()),
      user: {
        id: uuidv4(),
        name: "John",
        photo: "https://randomuser.me/api/portraits/men/1.jpg",
      },
    },
    {
      id: uuidv4(),
      text: "I'm good thanks! 😊",
      time: FormatTime(new Date()),
      user: {
        id: uuidv4(),
        name: "Sarah",
        photo: "https://randomuser.me/api/portraits/women/1.jpg",
      },
    },
  ]);

  useEffect(() => {
    const connectWebSocket = () => {
      wsRef.current = new WebSocket(
        `ws://localhost:3002/video-conference/ws/${roomID}/${userID}`,
      );

      wsRef.current.onopen = () => {
        console.log("WebSocket connection established");
        wsRef.current?.send(
          JSON.stringify({
            type: "user-joined",
            userID,
            userName,
            userPhoto,
          }),
        );
      };

      wsRef.current.onmessage = async (event) => {
        const message = JSON.parse(event.data);

        switch (message.type) {
          case "user-joined":
            console.log("User joined:", message);
            if (message.userID !== userID) {
              setUsers((prevUsers) => [
                ...prevUsers,
                {
                  id: message.userID,
                  name: message.userName,
                  imgUrl: message.userPhoto,
                  isAudioOn: true,
                  isPinned: false,
                  isVideoOn: true,
                  videoStream: null,
                },
              ]);
              initiateWebRTCConnection(message.userID);
            }
            break;

          case "user-left":
            setUsers((prevUsers) =>
              prevUsers.filter((user) => user.id !== message.userID),
            );
            removeRemoteStream(message.userID);
            closePeerConnection(message.userID);
            break;

          case "users-list":
            setUsers(message.users);
            message.users.forEach((user: { id: string }) => {
              if (user.id !== userID) {
                initiateWebRTCConnection(user.id);
              }
            });
            break;

          case "offer":
            await handleOffer(message.offer, message.from);
            break;

          case "answer":
            await handleAnswer(message.answer, message.from);
            break;

          case "ice-candidate":
            await handleICECandidate(message.candidate, message.from);
            break;

          case "chat-message":
            setChatMessages((prevMessages) => [
              ...prevMessages,
              {
                id: message.id,
                text: message.text,
                time: new Date().toLocaleTimeString(),
                user: {
                  id: message.user.id,
                  name: message.user.name,
                  photo: message.user.photo,
                },
              },
            ]);
            break;

          default:
            console.log("Received message:", message);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      wsRef.current.onclose = () => {
        console.log("WebSocket connection closed. Reconnecting...");
        setTimeout(connectWebSocket, 3000);
      };
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      Object.values(peerConnections.current).forEach((pc) => pc.close());
    };
  }, [roomID, userID, userName, userPhoto, setUsers, removeRemoteStream]);

  const sendChatMessage = (text: string) => {
    if (wsRef.current) {
      const message: Message = {
        id: uuidv4(),
        text,
        time: new Date().toLocaleTimeString(),
        user: {
          id: userID,
          name: userName,
          photo: userPhoto,
        },
      };

      wsRef.current.send(
        JSON.stringify({
          type: "chat-message",
          message: message,
        }),
      );
    }
  };

  const initiateWebRTCConnection = async (remoteUserID: string) => {
    if (peerConnections.current[remoteUserID]) {
      console.log("Connection already exists with user:", remoteUserID);
      return;
    }

    const peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    localStream?.getTracks().forEach((track) => {
      peerConnection.addTrack(track, localStream);
    });

    peerConnection.ontrack = (event) => {
      const remoteStream = event.streams[0];
      addRemoteStream(remoteUserID, remoteStream);
    };

    peerConnection.onicecandidate = (event) => {
      if (event.candidate && wsRef.current) {
        wsRef.current.send(
          JSON.stringify({
            type: "ice-candidate",
            candidate: event.candidate,
            to: remoteUserID,
          }),
        );
      }
    };

    try {
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);
      if (wsRef.current) {
        wsRef.current.send(
          JSON.stringify({
            type: "offer",
            offer: offer,
            to: remoteUserID,
          }),
        );
      }
    } catch (error) {
      console.error("Error creating offer:", error);
    }

    peerConnections.current[remoteUserID] = peerConnection;
  };

  const handleOffer = async (
    offer: RTCSessionDescriptionInit,
    remoteUserID: string,
  ) => {
    if (peerConnections.current[remoteUserID]) {
      console.log("Connection already exists with user:", remoteUserID);
      return;
    }

    const peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    localStream?.getTracks().forEach((track) => {
      peerConnection.addTrack(track, localStream);
    });

    peerConnection.ontrack = (event) => {
      const remoteStream = event.streams[0];
      addRemoteStream(remoteUserID, remoteStream);
    };

    peerConnection.onicecandidate = (event) => {
      if (event.candidate && wsRef.current) {
        wsRef.current.send(
          JSON.stringify({
            type: "ice-candidate",
            candidate: event.candidate,
            to: remoteUserID,
          }),
        );
      }
    };

    try {
      await peerConnection.setRemoteDescription(offer);
      const answer = await peerConnection.createAnswer();
      await peerConnection.setLocalDescription(answer);
      if (wsRef.current) {
        wsRef.current.send(
          JSON.stringify({
            type: "answer",
            answer: answer,
            to: remoteUserID,
          }),
        );
      }
    } catch (error) {
      console.error("Error handling offer:", error);
    }

    peerConnections.current[remoteUserID] = peerConnection;
  };

  const handleAnswer = async (
    answer: RTCSessionDescriptionInit,
    remoteUserID: string,
  ) => {
    const peerConnection = peerConnections.current[remoteUserID];
    if (peerConnection) {
      try {
        await peerConnection.setRemoteDescription(answer);
      } catch (error) {
        console.error("Error setting remote description:", error);
      }
    }
  };

  const handleICECandidate = async (
    candidate: RTCIceCandidateInit,
    remoteUserID: string,
  ) => {
    const peerConnection = peerConnections.current[remoteUserID];
    if (peerConnection) {
      try {
        await peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
      } catch (error) {
        console.error("Error adding ICE candidate:", error);
      }
    }
  };

  const closePeerConnection = (remoteUserID: string) => {
    const peerConnection = peerConnections.current[remoteUserID];
    if (peerConnection) {
      peerConnection.close();
      delete peerConnections.current[remoteUserID];
    }
  };

  return {
    chatMessages,
    sendChatMessage,
  };
};
