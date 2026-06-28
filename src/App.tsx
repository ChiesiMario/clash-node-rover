
import * as Tabs from "@radix-ui/react-tabs";
import { Activity, Settings2, TerminalSquare } from "lucide-react";
import { Dashboard, AppStatus } from "./components/Dashboard";
import { Settings } from "./components/Settings";
import { Console } from "./components/Console";
import "./App.css";
import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";

function App() {
  const [status, setStatus] = useState<AppStatus | null>(null);
  const [activeTab, setActiveTab] = useState("dashboard");

  useEffect(() => {
    invoke<AppStatus>("get_status").then((initialStatus) => {
      setStatus(initialStatus);
    }).catch(console.error);

    const unlisten = listen<AppStatus>("status_update", (event) => {
      setStatus(event.payload);
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);
  return (
    <div className="flex flex-col h-screen bg-background text-foreground selection:bg-primary selection:text-primary-foreground">
      <Tabs.Root value={activeTab} onValueChange={setActiveTab} className="flex flex-col h-full">
        <Tabs.List className="flex border-b border-border bg-muted/30 px-4 py-2 gap-1">
          <Tabs.Trigger
            value="dashboard"
            className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
          >
            <Activity className="w-4 h-4" />
            <span className="font-medium text-sm">Dashboard</span>
          </Tabs.Trigger>
          <Tabs.Trigger
            value="settings"
            className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
          >
            <Settings2 className="w-4 h-4" />
            <span className="font-medium text-sm">Settings</span>
          </Tabs.Trigger>
          <Tabs.Trigger
            value="console"
            className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-all data-[state=active]:bg-primary/10 data-[state=active]:text-primary"
          >
            <TerminalSquare className="w-4 h-4" />
            <span className="font-medium text-sm">Console</span>
          </Tabs.Trigger>
        </Tabs.List>

        <div className="flex-1 overflow-auto">
          <Tabs.Content 
            value="dashboard" 
            className="h-full focus:outline-none data-[state=inactive]:hidden animate-in fade-in duration-1000" 
            forceMount
          >
            <Dashboard status={status} onNavigate={setActiveTab} />
          </Tabs.Content>
          <Tabs.Content 
            value="settings" 
            className="h-full focus:outline-none data-[state=inactive]:hidden animate-in fade-in duration-1000" 
            forceMount
          >
            <Settings />
          </Tabs.Content>
          <Tabs.Content 
            value="console" 
            className="h-full focus:outline-none data-[state=inactive]:hidden animate-in fade-in duration-1000" 
            forceMount
          >
            <Console />
          </Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  );
}

export default App;
