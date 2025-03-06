import "./index.css";
import "bootstrap/dist/css/bootstrap.min.css";
import { Home } from "./Components/Home";
import { SidebarProvider } from "./Contexts/SidebarContext.tsx";

export const App = () => {
  return (
    <SidebarProvider>
      <Home />
    </SidebarProvider>
  );
};
