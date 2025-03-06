import { SideBarButton } from "./SideBarButton";
import { Home, MessageCircle, Settings } from "lucide-react";
import { useState } from "react";

export const SideBar = () => {
  const [selectedItem, setSelectedItem] = useState("Home");

  const items = [
    {
      name: "Home",
      Icon: Home,
      action: () => console.log("Navigating to home..."),
    },
    {
      name: "Messages",
      Icon: MessageCircle,
      action: () => console.log("Navigating to messages..."),
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
    <div className="flex flex-col gap-2 my-5">
      {items.map((item) => (
        <SideBarButton
          key={item.name}
          onClick={() => handleClick(item.name)}
          variant={item.name === selectedItem ? "default" : "dark"}
          className="!rounded-full font-bold p-3 flex flex-col items-center"
        >
          <item.Icon className="w-6 h-6 mb-2 !stroke-2" />
          <span className="text-sm">{item.name}</span>
        </SideBarButton>
      ))}
    </div>
  );
};
