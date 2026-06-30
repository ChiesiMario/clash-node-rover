import { useTranslation } from "react-i18next";
import { getVersion } from "@tauri-apps/api/app";
import { openUrl } from "@tauri-apps/plugin-opener";
import { useEffect, useState } from "react";
import { ExternalLink, User, Info } from "lucide-react";

export function About() {
  const { t } = useTranslation();
  const [version, setVersion] = useState<string>("");

  useEffect(() => {
    getVersion().then(setVersion).catch(console.error);
  }, []);

  const openGithub = () => {
    openUrl("https://github.com/ChiesiMario/clash-node-rover").catch(console.error);
  };

  return (
    <div className="h-full w-full flex flex-col items-center justify-center p-8 bg-background relative overflow-hidden">
      {/* Background decorations */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/5 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl pointer-events-none" />
      
      <div className="z-10 flex flex-col items-center max-w-xl w-full">
        {/* Logo and Title Section */}
        <div className="flex flex-col items-center mb-10 space-y-6">
          <img src="/logo.svg" alt="Clash Node Rover" className="w-32 h-32 drop-shadow-lg" />
          <div className="text-center space-y-2">
            <h1 className="text-4xl font-bold tracking-tight text-foreground">
              {t('about.title', 'About Clash Node Rover')}
            </h1>
            <p className="text-lg text-muted-foreground max-w-md mx-auto">
              {t('about.description', 'Lightweight and smart node management tool')}
            </p>
          </div>
        </div>

        {/* Info Cards Section */}
        <div className="w-full grid grid-cols-1 md:grid-cols-2 gap-4 mb-10">
          <div className="flex items-center p-4 bg-card rounded-2xl border border-border shadow-sm hover:shadow-md transition-shadow">
            <div className="p-3 bg-primary/10 text-primary rounded-xl mr-4">
              <Info className="w-6 h-6" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('about.version', 'Version')}</p>
              <p className="font-semibold text-foreground text-lg">{version || "..."}</p>
            </div>
          </div>
          
          <div className="flex items-center p-4 bg-card rounded-2xl border border-border shadow-sm hover:shadow-md transition-shadow">
            <div className="p-3 bg-primary/10 text-primary rounded-xl mr-4">
              <User className="w-6 h-6" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('about.author', 'Author')}</p>
              <p className="font-semibold text-foreground text-lg">ChiesiMario</p>
            </div>
          </div>
        </div>

        {/* Action Button */}
        <button
          onClick={openGithub}
          className="flex items-center gap-3 px-8 py-4 bg-foreground text-background rounded-full hover:bg-foreground/90 transition-all shadow-lg hover:shadow-xl hover:-translate-y-1 active:translate-y-0"
        >
          <ExternalLink className="w-6 h-6" />
          <span className="font-medium text-lg">{t('about.github', 'GitHub Repository')}</span>
        </button>
      </div>
    </div>
  );
}
