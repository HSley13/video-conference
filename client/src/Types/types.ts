export type UserInfo = {
  id: string;
  username: string;
  email: string;
  imageUrl: string;
};

export type Message = {
  id: string;
  text: string;
  time: string;
  user: {
    id: string;
    name: string;
    imgUrl: string;
  };
};

export type User = {
  id: string;
  name: string;
  imgUrl: string;
  isAudioOn: boolean;
  isPinned: boolean;
  videoStream: MediaStream | null;
  isVideoOn?: boolean;
};
