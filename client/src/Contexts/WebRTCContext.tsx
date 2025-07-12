import {
  createContext,
  useContext,
  useState,
  useRef,
  useEffect,
  useCallback,
  Dispatch,
  SetStateAction,
  ReactNode,
  useMemo,
} from "react";
import { useSearchParams } from "react-router-dom";

import { useAsync } from "../Hooks/useAsync";
import { getUserInfo } from "../Services/user";
import { User, UserInfo } from "../Types/types";
import { useUser } from "../Hooks/useUser";

type RemoteStreamMap = Record<string, MediaStream>;

export interface WebRTCContextValue {
  localStream: MediaStream | null;
  remoteStreams: RemoteStreamMap;
  users: User[];
  setUsers: Dispatch<SetStateAction<User[]>>;
  addRemoteStream: (uid: string, s: MediaStream) => void;
  removeRemoteStream: (uid: string) => void;
  roomId: string;
  userInfo: UserInfo;
}

const WebRTCContext = createContext<WebRTCContextValue | undefined>(undefined);

export const WebRTCProvider = ({ children }: { children: ReactNode }) => {
  const [search] = useSearchParams();
  const roomId = search.get("room") ?? "";
  const user = useUser();
  const userIdParam = search.get("uid") ?? user?.id;
  console.log("userIdParam:", userIdParam);
  console.log("roomId:", roomId);

  const [userInfo, setUserInfo] = useState<UserInfo>({
    id: userIdParam ?? "",
    userName: "",
    email: "",
    imgUrl: "",
  });

  const fetchProfile = useCallback(
    () =>
      userIdParam ? getUserInfo({ id: userIdParam }) : Promise.resolve(null),
    [userIdParam],
  );

  const { value: fetched } = useAsync(fetchProfile, [fetchProfile]);

  useEffect(() => {
    if (!fetched) return;
    setUserInfo({
      id: fetched.id,
      userName: fetched.userName,
      email: fetched.email,
      imgUrl: fetched.imgUrl,
    });
  }, [fetched]);

  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<RemoteStreamMap>({});
  const [users, setUsers] = useState<User[]>([]);
  const mediaRequestedRef = useRef(false);

  const addRemoteStream = useCallback((uid: string, stream: MediaStream) => {
    setRemoteStreams((prev) => ({ ...prev, [uid]: stream }));
    setUsers((prev) =>
      prev.map((u) => (u.id === uid ? { ...u, videoStream: stream } : u)),
    );
  }, []);

  const removeRemoteStream = useCallback((uid: string) => {
    setRemoteStreams((prev) => {
      const { [uid]: gone, ...rest } = prev;
      gone?.getTracks().forEach((t) => t.stop());
      return rest;
    });
    setUsers((prev) =>
      prev.map((u) => (u.id === uid ? { ...u, videoStream: null } : u)),
    );
  }, []);

  useEffect(() => {
    if (mediaRequestedRef.current) return;
    mediaRequestedRef.current = true;

    (async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: true,
        });
        setLocalStream(stream);
      } catch (e) {
        console.error("[WebRTC] getUserMedia:", e);
      }
    })();

    return () => {
      setLocalStream((s) => {
        s?.getTracks().forEach((t) => t.stop());
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

  const ctxValue = useMemo<WebRTCContextValue>(
    () => ({
      localStream,
      remoteStreams,
      users,
      addRemoteStream,
      removeRemoteStream,
      setUsers,
      roomId,
      userInfo,
    }),
    [
      localStream,
      remoteStreams,
      users,
      addRemoteStream,
      removeRemoteStream,
      roomId,
      userInfo,
    ],
  );

  return (
    <WebRTCContext.Provider value={ctxValue}>{children}</WebRTCContext.Provider>
  );
};

export const useWebRTC = (): WebRTCContextValue => {
  const ctx = useContext(WebRTCContext);
  if (!ctx) throw new Error("useWebRTC must be used inside <WebRTCProvider>");
  return ctx;
};
