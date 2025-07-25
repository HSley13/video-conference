import { useEffect, useRef, useState, useCallback } from "react";
import { v4 as uuid } from "uuid";
import { useWebRTC } from "../Contexts/WebRTCContext";
import { Message, User } from "../Types/types";
import { FormatTime } from "../Utils/utils";

interface WsUser {
  userID: string;
  userName: string;
  imgUrl: string;
}
interface UsersListMsg {
  type: "users-list";
  users: WsUser[];
}
interface UserJoinedMsg {
  type: "user-joined";
  userID: string;
  userName: string;
  imgUrl: string;
}
interface UserLeftMsg {
  type: "user-left";
  userID: string;
}
interface OfferMsg {
  type: "offer";
  from: string;
  offer: RTCSessionDescriptionInit;
}
interface AnswerMsg {
  type: "answer";
  from: string;
  answer: RTCSessionDescriptionInit;
}
interface IceCandidateMsg {
  type: "ice-candidate";
  from: string;
  candidate: RTCIceCandidateInit;
}
interface ChatMsg {
  type: "chat-message";
  id: string;
  text: string;
  time: string;
  user: Pick<User, "id" | "userName" | "imgUrl">;
}

type Incoming =
  | UsersListMsg
  | UserJoinedMsg
  | UserLeftMsg
  | OfferMsg
  | AnswerMsg
  | IceCandidateMsg
  | ChatMsg;

type Outgoing =
  | {
      type: "offer";
      offer: RTCSessionDescriptionInit;
      to: string;
      from: string;
    }
  | {
      type: "answer";
      answer: RTCSessionDescriptionInit;
      to: string;
      from: string;
    }
  | {
      type: "ice-candidate";
      candidate: RTCIceCandidateInit;
      to: string;
      from: string;
    }
  | { type: "chat-message"; message: Message };

export const useVideoConference = () => {
  const {
    roomId,
    userInfo: { id: userID, userName, imgUrl },
    localStream,
    setLocalStream,
    addRemoteStream,
    removeRemoteStream,
    setUsers,
    peersRef,
  } = useWebRTC();

  const socketRef = useRef<WebSocket | null>(null);

  const [messages, setMessages] = useState<Message[]>([]);
  const seenMsgIds = useRef(new Set<string>());

  const knownUserIds = useRef<Set<string>>(new Set());

  const [isScreenSharing, setIsScreenSharing] = useState(false);

  useEffect(() => {
    if (!userID) return;

    const url = `ws://localhost:3002/video-conference/ws/${roomId}`;

    const openSocket = () => {
      const ws = new WebSocket(url);
      socketRef.current = ws;

      ws.onmessage = (e) => handleSocketMessage(JSON.parse(e.data) as Incoming);

      ws.onclose = () => {
        console.warn("[WS] closed – retry in 3s");
        setTimeout(openSocket, 3_000);
      };
    };

    openSocket();

    return () => {
      socketRef.current?.close();
      Object.values(peersRef.current).forEach((pc) => pc.close());
      peersRef.current = {};
      knownUserIds.current.clear();
    };
  }, [roomId, userID]);

  const handleSocketMessage = async (msg: Incoming) => {
    switch (msg.type) {
      case "users-list": {
        const remotes = msg.users.filter((u) => u.userID !== userID);
        knownUserIds.current = new Set(remotes.map((u) => u.userID));

        setUsers(
          remotes.map<User>((u) => ({
            id: u.userID,
            userName: u.userName,
            imgUrl: u.imgUrl,
            isAudioOn: true,
            isVideoOn: true,
            isPinned: false,
            videoStream: null,
          })),
        );

        if (localStream) {
          remotes.forEach((u) => createPeer(u.userID));
        }
        break;
      }

      case "user-joined": {
        if (msg.userID === userID || knownUserIds.current.has(msg.userID)) {
          break;
        }

        knownUserIds.current.add(msg.userID);

        setUsers((prev) => {
          if (prev.some((u) => u.id === msg.userID)) return prev;
          return [
            ...prev,
            {
              id: msg.userID,
              userName: msg.userName,
              imgUrl: msg.imgUrl,
              isAudioOn: true,
              isVideoOn: true,
              isPinned: false,
              videoStream: null,
            },
          ];
        });

        break;
      }

      case "user-left": {
        knownUserIds.current.delete(msg.userID);
        setUsers((prev) => prev.filter((u) => u.id !== msg.userID));
        removeRemoteStream(msg.userID);
        closePeer(msg.userID);
        break;
      }

      case "offer":
        if (msg.from !== userID) await handleOffer(msg.offer, msg.from);
        break;

      case "answer":
        if (msg.from !== userID) await handleAnswer(msg.answer, msg.from);
        break;

      case "ice-candidate":
        if (msg.from !== userID) await handleCandidate(msg.candidate, msg.from);
        break;

      case "chat-message":
        if (!seenMsgIds.current.has(msg.id)) {
          seenMsgIds.current.add(msg.id);
          setMessages((prev) => [
            ...prev,
            {
              id: msg.id,
              text: msg.text,
              time: msg.time,
              user: msg.user,
            },
          ]);
        }
        break;
    }
  };

  const sendChatMessage = (text: string) => {
    if (!socketRef.current || !text.trim()) return;
    const msg: Message = {
      id: uuid(),
      text,
      time: FormatTime(new Date()),
      user: { id: userID, userName, imgUrl },
    };
    seenMsgIds.current.add(msg.id);
    setMessages((prev) => [...prev, msg]);
    socketRef.current.send(
      JSON.stringify({ type: "chat-message", message: msg } satisfies Outgoing),
    );
  };

  const createPeer = async (remoteId: string) => {
    if (remoteId === userID || peersRef.current[remoteId]) return;

    const pc = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    localStream?.getTracks().forEach((t) => pc.addTrack(t, localStream));

    pc.ontrack = (e) => addRemoteStream(remoteId, e.streams[0]);

    pc.onicecandidate = (e) => {
      if (e.candidate && socketRef.current) {
        socketRef.current.send(
          JSON.stringify({
            type: "ice-candidate",
            candidate: e.candidate,
            to: remoteId,
            from: userID,
          } satisfies Outgoing),
        );
      }
    };

    await pc.setLocalDescription(await pc.createOffer());
    socketRef.current?.send(
      JSON.stringify({
        type: "offer",
        offer: pc.localDescription!,
        to: remoteId,
        from: userID,
      } satisfies Outgoing),
    );

    peersRef.current[remoteId] = pc;
  };

  const handleOffer = async (
    offer: RTCSessionDescriptionInit,
    remoteId: string,
  ) => {
    let pc = peersRef.current[remoteId];
    if (!pc) {
      pc = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
      });
      localStream?.getTracks().forEach((t) => pc.addTrack(t, localStream));
      pc.ontrack = (e) => addRemoteStream(remoteId, e.streams[0]);
      pc.onicecandidate = (e) =>
        e.candidate &&
        socketRef.current?.send(
          JSON.stringify({
            type: "ice-candidate",
            candidate: e.candidate,
            to: remoteId,
            from: userID,
          } satisfies Outgoing),
        );
      peersRef.current[remoteId] = pc;
    }

    if (pc.signalingState === "have-local-offer") {
      await pc.setLocalDescription({ type: "rollback" } as any);
    }

    await pc.setRemoteDescription(offer);
    await pc.setLocalDescription(await pc.createAnswer());

    socketRef.current?.send(
      JSON.stringify({
        type: "answer",
        answer: pc.localDescription!,
        to: remoteId,
        from: userID,
      } satisfies Outgoing),
    );
  };

  const handleAnswer = async (
    answer: RTCSessionDescriptionInit,
    remoteId: string,
  ) => {
    await peersRef.current[remoteId]?.setRemoteDescription(answer);
  };

  const handleCandidate = async (
    candidate: RTCIceCandidateInit,
    remoteId: string,
  ) => {
    await peersRef.current[remoteId]?.addIceCandidate(
      new RTCIceCandidate(candidate),
    );
  };

  const closePeer = (remoteId: string) => {
    peersRef.current[remoteId]?.close();
    delete peersRef.current[remoteId];
  };

  const renegotiateAll = useCallback(() => {
    Object.entries(peersRef.current).forEach(([id, pc]) => {
      if (pc.connectionState === "connected") {
        pc.createOffer()
          .then((o) => pc.setLocalDescription(o))
          .then(() =>
            socketRef.current?.send(
              JSON.stringify({
                type: "offer",
                offer: pc.localDescription!,
                to: id,
                from: userID,
              } satisfies Outgoing),
            ),
          );
      }
    });
  }, [userID]);

  const toggleScreenSharing = useCallback(async () => {
    const newStream = !isScreenSharing
      ? await navigator.mediaDevices.getDisplayMedia({ video: true })
      : await navigator.mediaDevices.getUserMedia({ video: true, audio: true });

    const newTrack = newStream.getVideoTracks()[0];
    if (!newTrack) return;

    Object.values(peersRef.current).forEach((pc) => {
      const sender = pc.getSenders().find((s) => s.track?.kind === "video");
      if (sender) sender.replaceTrack(newTrack);
    });

    setLocalStream((prev) => {
      prev?.getTracks().forEach((t) => t.stop());
      return newStream;
    });

    setIsScreenSharing((v) => !v);
    renegotiateAll();
  }, [isScreenSharing, setLocalStream, renegotiateAll]);

  return { messages, sendChatMessage, toggleScreenSharing, isScreenSharing };
};
