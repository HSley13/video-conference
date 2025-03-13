import { useEffect, useRef, useState } from "react";
import { useWebRTC } from "../Contexts/WebRTCContext";
import { Message } from "../Types/types";
import { FormatTime } from "../Utils/utils";
import { User } from "../Types/types";

export const useVideoConference = (
  roomID: string,
  userID: number,
  userName: string,
  userPhoto: string,
) => {
  const { localStream, addRemoteStream, removeRemoteStream, setUsers } =
    useWebRTC();
  const wsRef = useRef<WebSocket | null>(null);
  const peerConnections = useRef<Record<number, RTCPeerConnection>>({});

  const [chatMessages, setChatMessages] = useState<Message[]>([
    {
      id: 1,
      text: "Hey, how are you?",
      time: FormatTime(new Date()),
      user: {
        id: 1,
        name: "John",
        photo: "https://randomuser.me/api/portraits/men/1.jpg",
      },
    },
    {
      id: 2,
      text: "I'm good thanks! 😊",
      time: FormatTime(new Date()),
      user: {
        id: 2,
        name: "Sarah",
        photo: "https://randomuser.me/api/portraits/women/1.jpg",
      },
    },
  ]);

  useEffect(() => {
    const connectWebSocket = () => {
      wsRef.current = new WebSocket(
        `ws://localhost:8080/video-conference/ws/${roomID}`,
      );

      wsRef.current.onopen = () => {
        console.log("WebSocket connection established");
      };

      wsRef.current.onmessage = async (event) => {
        const message = JSON.parse(event.data);

        switch (message.type) {
          case "user-joined":
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
            break;

          case "user-left":
            setUsers((prevUsers) =>
              prevUsers.filter((user) => user.id !== message.userID),
            );
            removeRemoteStream(message.userID);
            break;

          case "users-list":
            setUsers(message.users);
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
    };
  }, [roomID, userID, setUsers, removeRemoteStream]);

  const sendChatMessage = (text: string) => {
    if (wsRef.current) {
      const message: Message = {
        id: Date.now(),
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

  const initiateWebRTCConnection = async (remoteUserID: number) => {
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

    peerConnections.current[remoteUserID] = peerConnection;
  };

  const handleOffer = async (
    offer: RTCSessionDescriptionInit,
    remoteUserID: number,
  ) => {
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

    peerConnections.current[remoteUserID] = peerConnection;
  };

  const handleAnswer = async (
    answer: RTCSessionDescriptionInit,
    remoteUserID: number,
  ) => {
    const peerConnection = peerConnections.current[remoteUserID];
    if (peerConnection) {
      await peerConnection.setRemoteDescription(answer);
    }
  };

  const handleICECandidate = async (
    candidate: RTCIceCandidateInit,
    remoteUserID: number,
  ) => {
    const peerConnection = peerConnections.current[remoteUserID];
    if (peerConnection) {
      await peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
    }
  };

  return {
    chatMessages,
    sendChatMessage,
  };
};
