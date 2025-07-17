import { useState, Fragment } from "react";
import { ScreenShare, MonitorX } from "lucide-react";
import { VideoCard } from "./VideoCard";
import { useWebRTC } from "../../Contexts/WebRTCContext";
import { useVideoConference } from "../../Hooks/websocket";

export const VideoList = () => {
  const { users, remoteStreams } = useWebRTC();

  const { toggleScreenSharing, isScreenSharing } = useVideoConference();

  const [pinnedId, setPinnedId] = useState<string | null>(null);

  const mainUser = pinnedId
    ? (users.find((u) => u.id === pinnedId) ?? null)
    : null;

  const otherUsers = mainUser
    ? users.filter((u) => u.id !== mainUser.id)
    : users;

  const streamFor = (uid: string) => remoteStreams[uid] ?? null;

  return (
    <div className="flex flex-col h-screen">
      <div className="flex items-center gap-4 p-2 bg-gray-700 text-white">
        <button
          className="flex items-center gap-1 px-3 py-1 rounded hover:bg-gray-600 transition-colors"
          onClick={() => void toggleScreenSharing()}
        >
          {isScreenSharing ? (
            <Fragment>
              <MonitorX size={16} />
              Stop share
            </Fragment>
          ) : (
            <Fragment>
              <ScreenShare size={16} />
              Share screen
            </Fragment>
          )}
        </button>
      </div>

      <div className="flex-1 flex flex-col bg-gray-200 rounded-lg p-2">
        <div className="flex-1 w-full rounded-xl overflow-hidden mb-4">
          {mainUser ? (
            <VideoCard
              key={mainUser.id}
              {...mainUser}
              videoStream={streamFor(mainUser.id)}
              onPin={setPinnedId}
            />
          ) : (
            <div className="w-full h-full bg-gray-300 flex items-center justify-center text-gray-500 rounded-xl">
              Click the pin icon on any participant to feature them here
            </div>
          )}
        </div>

        <div className="h-40 bg-gray-200 rounded-xl">
          <div className="flex gap-4 overflow-x-auto h-full px-2">
            {otherUsers.map((user) => (
              <div
                key={user.id}
                className="flex-shrink-0 w-44 h-full last:pr-2"
              >
                <VideoCard
                  {...user}
                  videoStream={streamFor(user.id)}
                  onPin={setPinnedId}
                />
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
