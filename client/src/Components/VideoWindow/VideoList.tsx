import { useState } from "react";
import { VideoCard } from "./VideoCard";
import { useWebRTC } from "../../Contexts/WebRTCContext";

export const VideoList = () => {
  // const { localStream, remoteStreams, users } = useWebRTC();
  //
  // const usersWithStream = users.map((user) => ({
  //   ...user,
  //   videoStream: remoteStreams[user.id],
  // }));

  // const handlePin = (userId: number) => {
  //   setUsers((prevUsers) =>
  //     prevUsers.map((user) => ({
  //       ...user,
  //       isPinned: user.id === userId ? !user.isPinned : false,
  //     })),
  //   );
  // };
  // const pinnedUser = users.find((user) => user.isPinned);
  // const mainUser = pinnedUser || users[0];
  // const otherUsers = usersWithStream.filter((user) => user.id !== mainUser.id);
  const { users } = useWebRTC();
  const [pinnedId, setPinnedId] = useState<number | null>(1);
  const mainUser = pinnedId ? users.find((user) => user.id === pinnedId) : null;
  const otherUsers = users.filter((user) => user.id !== pinnedId);

  return (
    <div className="flex flex-col h-screen w-100 bg-gray-200 rounded-lg">
      <div className="flex-1 w-100 rounded-xl overflow-hidden">
        {mainUser ? (
          <VideoCard key={mainUser.id} {...mainUser} onPin={setPinnedId} />
        ) : (
          <div className="w-100 h-100 bg-gray-200 flex items-center justify-center text-gray-400">
            Click the pin icon on any participant to feature them here
          </div>
        )}
      </div>

      <div className="h-40 mt-4 bg-gray-200 rounded-xl">
        <div className="flex gap-4 overflow-x-auto h-full">
          {otherUsers.map((user) => (
            <div key={user.id} className="flex-shrink-0 w-45 h-full">
              <VideoCard key={user.id} {...user} onPin={setPinnedId} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
