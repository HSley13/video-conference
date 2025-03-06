import { SideBarButton } from "./SideBarButton";
import { Home, Settings } from "lucide-react";
import { useState } from "react";
import { Button } from "react-bootstrap";
import { Link } from "react-router-dom";

export const SideBar = () => {
  const [selectedItem, setSelectedItem] = useState("Home");

  const items = [
    {
      name: "Home",
      Icon: Home,
      action: () => console.log("Navigating to home..."),
    },
    {
      name: "Settings",
      Icon: Settings,
      action: () => console.log("Navigating to settings..."),
    },
  ];

  const handleClick = (itemName: string) => {
    setSelectedItem(itemName);
    const item = items.find((i) => i.name === itemName);
    item?.action();
  };

  return (
    <aside className="border-r-3 bg-gray-200 border-gray-200 h-full flex flex-col ">
      <div className="flex flex-col m-2 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
        <div className="border-b-2 border-gray-400 p-2 m-2 flex flex-col items-center">
          <Link to="/" className="flex flex-col items-center w-full">
            <Button className="bg-transparent p-0 border-0 flex flex-col items-center">
              <img
                src="https://img.posterstore.com/zoom/wb0125-8batman-portrait50x70-34329-40892.jpg"
                alt="Profile"
                className="w-15 h-15 rounded-full mx-auto"
              />
            </Button>
          </Link>
          <span className="text-sm font-bold mt-1">Profile</span>
        </div>

        {items.map((item) => (
          <SideBarButton
            key={item.name}
            onClick={() => handleClick(item.name)}
            variant={item.name === selectedItem ? "default" : "dark"}
            className="!rounded-full font-bold p-3 flex flex-col items-center flex-shrink-0"
          >
            <item.Icon className="w-6 h-6 mb-2 !stroke-2" />
            <span className="text-sm">{item.name}</span>
          </SideBarButton>
        ))}
      </div>
    </aside>
  );
};
