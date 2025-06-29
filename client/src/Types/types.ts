export type UserInfo = {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  imageUrl: string;
  newPassword?: string;
  confirmPassword?: string;
  accessToken?: string;
  refreshToken?: string;
};

export type Message = {
  id: string;
  text: string;
  time: string;
  user: {
    id: string;
    name: string;
    photo: string;
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
