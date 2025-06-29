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
} from "react";
import { getUserInfo } from "../Services/auth";
import { useAsync } from "../Hooks/useAsync";
import { useUser } from "../Hooks/useUser";
import { User, UserInfo } from "../Types/types";

const qs = new URLSearchParams(window.location.search);
const envTokenRaw = import.meta.env.VITE_ACCESS_TOKEN as string | undefined;
const accessToken =
  envTokenRaw?.trim() || localStorage.getItem("access_token") || "";

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

export function WebRTCProvider({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  const roomId = qs.get("room") ?? "550e8400-e29b-41d4-a716-446655440000";

  const fallbackId = useUser();
  const userIdParam =
    qs.get("uid") ?? import.meta.env.VITE_USER_ID ?? fallbackId;
  const firstName =
    qs.get("firstName") ?? import.meta.env.VITE_FIRST_NAME ?? "";
  const imageUrlQS = qs.get("imageUrl") ?? import.meta.env.VITE_IMAGE_URL ?? "";
  const emailQS = qs.get("email") ?? import.meta.env.VITE_EMAIL ?? "";

  const [userInfo, setUserInfo] = useState<UserInfo>({
    id: userIdParam,
    firstName,
    lastName: "",
    email: emailQS,
    imageUrl: imageUrlQS,
    newPassword: "",
    confirmPassword: "",
    accessToken,
  });

  const { value: fetched } = useAsync(
    () => getUserInfo({ id: userIdParam ?? "" }),
    [userIdParam],
  );

  useEffect(() => {
    if (!fetched) return;
    setUserInfo((prev) => ({
      ...prev,
      accessToken: fetched.accessToken,
      newPassword: fetched.newPassword,
      congirmPassword: fetched.confirmPassword,
    }));
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

  const ctxValue: WebRTCContextValue = {
    localStream,
    remoteStreams,
    users,
    addRemoteStream,
    removeRemoteStream,
    setUsers,
    roomId,
    userInfo,
  };

  return (
    <WebRTCContext.Provider value={ctxValue}>{children}</WebRTCContext.Provider>
  );
}

export const useWebRTC = (): WebRTCContextValue => {
  const ctx = useContext(WebRTCContext);
  if (!ctx) throw new Error("useWebRTC must be used inside <WebRTCProvider>");
  return ctx;
};
