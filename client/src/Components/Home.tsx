import { SideBar } from "./SideBar/SideBar";
import { MainWindow } from "./VideoWindow/MainWindow";

export const Home = () => {
  return (
    <div className="flex h-screen bg-gray-100">
      <div className="fixed top-0 left-0 bottom-0">
        <SideBar />
      </div>

      <div>
        <MainWindow />
      </div>
    </div>
  );
};
