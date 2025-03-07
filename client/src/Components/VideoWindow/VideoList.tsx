import { useState } from "react";
import { VideoCard } from "./VideoCard";
import { User } from "../../Types/types";

export const VideoList = () => {
  const [pinnedId, setPinnedId] = useState<number | null>(null);
  const [users] = useState<User[]>([
    {
      id: 1,
      name: "John",
      photo: "https://randomuser.me/api/portraits/men/1.jpg",
      isAudioOn: true,
      videoStream: null,
    },
    {
      id: 2,
      name: "Sarah",
      photo: "https://randomuser.me/api/portraits/women/1.jpg",
      isAudioOn: false,
      videoStream: null,
    },
    {
      id: 3,
      name: "Alex",
      photo: "https://randomuser.me/api/portraits/men/2.jpg",
      isAudioOn: true,
      videoStream: null,
    },
    {
      id: 4,
      name: "Maria",
      photo: "https://randomuser.me/api/portraits/women/2.jpg",
      isAudioOn: true,
      videoStream: null,
    },
    {
      id: 5,
      name: "John",
      photo: "https://randomuser.me/api/portraits/men/1.jpg",
      isAudioOn: false,
      videoStream: null,
    },
  ]);

  const mainUser = pinnedId ? users.find((user) => user.id === pinnedId) : null;
  const otherUsers = users.filter((user) => user.id !== pinnedId);

  return (
    <div className="flex flex-col h-screen bg-gray-900 p-4">
      <div className="flex-1 mb-4 rounded-xl overflow-hidden">
        {mainUser ? (
          <VideoCard
            key={mainUser.id}
            id={mainUser.id}
            imgUrl={mainUser.photo}
            name={mainUser.name}
            isPinned={true}
            videoStream={mainUser.videoStream}
            onPin={setPinnedId}
          />
        ) : (
          <div className="w-full h-full bg-gray-800 flex items-center justify-center text-gray-400">
            Click the pin icon on any participant to feature them here
          </div>
        )}
      </div>

      <div className="h-32 bg-gray-800 rounded-xl p-2">
        <div className="flex gap-4 overflow-x-auto h-full">
          {otherUsers.map((user) => (
            <div key={user.id} className="flex-shrink-0 w-48 h-full">
              <VideoCard
                id={user.id}
                name={user.name}
                imgUrl={user.photo}
                isPinned={false}
                videoStream={user.videoStream}
                onPin={setPinnedId}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
