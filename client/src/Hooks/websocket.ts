import { useEffect, useRef, useState } from "react";
import { useWebRTC } from "../Contexts/WebRTCContext";
import { Message } from "../Types/types";
import { FormatTime } from "../Utils/utils";
import { v4 as uuidv4 } from "uuid";

const getSubFromJWT = (jwt: string): string | null => {
  try {
    const [, payload] = jwt.split(".");
    return JSON.parse(atob(payload)).sub as string;
  } catch {
    return null;
  }
};

export const useVideoConference = (
  roomID: string,
  userIDInput: string | undefined,
  userName: string,
  userPhoto: string,
  accessToken: string,
) => {
  const userID = userIDInput ?? getSubFromJWT(accessToken) ?? "";
  const { localStream, addRemoteStream, removeRemoteStream, setUsers } =
    useWebRTC();

  const wsRef = useRef<WebSocket | null>(null);
  const peerConnections = useRef<Record<string, RTCPeerConnection>>({});
  const [chatMessages, setChatMessages] = useState<Message[]>([]);

  useEffect(() => {
    if (!userID || !accessToken) return;

    const connect = () => {
      wsRef.current = new WebSocket(
        `ws://localhost:3002/video-conference/ws/${roomID}/${userID}` +
          `?access_token=${encodeURIComponent(accessToken)}`,
      );

      wsRef.current.onopen = () => console.log("WebSocket open ✔");

      wsRef.current.onmessage = async (ev) => {
        const msg = JSON.parse(ev.data);

        switch (msg.type) {
          case "users-list":
            setUsers(
              msg.users.map((u: any) => ({
                id: u.id,
                name: u.name,
                imgUrl: u.imgUrl,
                isAudioOn: true,
                isVideoOn: true,
                isPinned: false,
                videoStream: null,
              })),
            );
            msg.users.forEach((u: any) => {
              if (u.id !== userID) initiateWebRTC(u.id);
            });
            break;

          case "user-joined":
            if (msg.userID !== userID) {
              setUsers((prev) => [
                ...prev,
                {
                  id: msg.userID,
                  name: msg.userName,
                  imgUrl: msg.userPhoto,
                  isAudioOn: true,
                  isVideoOn: true,
                  isPinned: false,
                  videoStream: null,
                },
              ]);
              initiateWebRTC(msg.userID);
            }
            break;

          case "user-left":
            setUsers((prev) => prev.filter((u) => u.id !== msg.userID));
            removeRemoteStream(msg.userID);
            closePeer(msg.userID);
            break;

          case "offer":
            await handleOffer(msg.offer, msg.from);
            break;

          case "answer":
            await handleAnswer(msg.answer, msg.from);
            break;

          case "ice-candidate":
            await handleCandidate(msg.candidate, msg.from);
            break;

          case "chat-message":
            setChatMessages((p) => [
              ...p,
              {
                id: msg.id,
                text: msg.text,
                time: msg.time,
                user: msg.user,
              },
            ]);
            break;

          default:
            console.log("unhandled:", msg);
        }
      };

      wsRef.current.onerror = (e) => console.error("WS error", e);
      wsRef.current.onclose = () => {
        console.log("WS closed → retry in 3 s");
        setTimeout(connect, 3_000);
      };
    };

    connect();
    return () => {
      wsRef.current?.close();
      Object.values(peerConnections.current).forEach((pc) => pc.close());
    };
  }, [roomID, userID, accessToken]);

  const sendChatMessage = (text: string) => {
    if (!wsRef.current) return;
    const message: Message = {
      id: uuidv4(),
      text,
      time: FormatTime(new Date()),
      user: { id: userID, name: userName, photo: userPhoto },
    };
    wsRef.current.send(JSON.stringify({ type: "chat-message", message }));
  };

  const initiateWebRTC = async (remoteID: string) => {
    if (peerConnections.current[remoteID]) return;

    const pc = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    localStream?.getTracks().forEach((t) => pc.addTrack(t, localStream));

    pc.ontrack = (ev) => addRemoteStream(remoteID, ev.streams[0]);
    pc.onicecandidate = (ev) =>
      ev.candidate &&
      wsRef.current?.send(
        JSON.stringify({
          type: "ice-candidate",
          candidate: ev.candidate,
          to: remoteID,
        }),
      );

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    wsRef.current?.send(JSON.stringify({ type: "offer", offer, to: remoteID }));

    peerConnections.current[remoteID] = pc;
  };

  const handleOffer = async (
    offer: RTCSessionDescriptionInit,
    remoteID: string,
  ) => {
    if (peerConnections.current[remoteID]) return;
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    localStream?.getTracks().forEach((t) => pc.addTrack(t, localStream));
    pc.ontrack = (ev) => addRemoteStream(remoteID, ev.streams[0]);
    pc.onicecandidate = (ev) =>
      ev.candidate &&
      wsRef.current?.send(
        JSON.stringify({
          type: "ice-candidate",
          candidate: ev.candidate,
          to: remoteID,
        }),
      );

    await pc.setRemoteDescription(offer);
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    wsRef.current?.send(
      JSON.stringify({ type: "answer", answer, to: remoteID }),
    );

    peerConnections.current[remoteID] = pc;
  };

  const handleAnswer = async (
    answer: RTCSessionDescriptionInit,
    remoteID: string,
  ) => {
    await peerConnections.current[remoteID]?.setRemoteDescription(answer);
  };

  const handleCandidate = async (
    cand: RTCIceCandidateInit,
    remoteID: string,
  ) => {
    await peerConnections.current[remoteID]?.addIceCandidate(
      new RTCIceCandidate(cand),
    );
  };

  const closePeer = (remoteID: string) => {
    peerConnections.current[remoteID]?.close();
    delete peerConnections.current[remoteID];
  };

  return { chatMessages, sendChatMessage };
};
