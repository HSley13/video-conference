export type Message = {
  id: number;
  text: string;
  time: string;
  user: {
    id: number;
    name: string;
    photo: string;
  };
};

export type User = {
  id: number;
  name: string;
  imgUrl: string;
  isAudioOn: boolean;
  isPinned: boolean;
  isVideoOn: boolean;
  videoStream: MediaStream | null;
};
