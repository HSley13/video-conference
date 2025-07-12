export type UserInfo = {
  id: string;
  userName: string;
  email: string;
  imgUrl: string;
};

export type Message = {
  id: string;
  text: string;
  time: string;
  user: {
    id: string;
    userName: string;
    imgUrl: string;
  };
};

export type User = {
  id: string;
  userName: string;
  imgUrl: string;
  isAudioOn: boolean;
  isPinned: boolean;
  videoStream: MediaStream | null;
  isVideoOn?: boolean;
};
