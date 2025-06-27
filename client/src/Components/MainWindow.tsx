import { SideBar } from "./SideBar/SideBar";
import { Home } from "./Home";

export const MainWindow = () => {
  return (
    <div className="flex h-screen bg-gray-100">
      <div className="fixed top-0 left-0 bottom-0">
        <SideBar />
      </div>

      <div className="flex-1 p-5">
        <Home />
      </div>
    </div>
  );
};
