export type Participant = {
  id: number;
  name: string;
  photo: string;
  isPinned: boolean;
  videoOn: boolean;
  audioOn: boolean;
};

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
