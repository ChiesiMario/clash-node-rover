import { useState, useEffect, useRef } from "react";
import { Zap, ChevronDown } from "lucide-react";
import { useTranslation } from "react-i18next";

export interface NodeResult {
  name: string;
  delay: number | null;
  mean?: number | null;
  jitter?: number;
  is_active: boolean;
  provider?: string;
  backoff_rounds?: number | null;
}

export const getColorClass = (delay: number | null, type: "bg" | "text") => {
  if (delay === null) return type === "bg" ? "bg-rose-600/80" : "text-rose-600 dark:text-rose-500";
  if (delay <= 150) return type === "bg" ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]" : "text-emerald-600 dark:text-emerald-400";
  if (delay <= 300) return type === "bg" ? "bg-amber-500" : "text-amber-600 dark:text-amber-500";
  if (delay <= 500) return type === "bg" ? "bg-orange-500" : "text-orange-600 dark:text-orange-500";
  return type === "bg" ? "bg-pink-500" : "text-pink-600 dark:text-pink-400";
};

export const getJitterColorClass = (jitter: number) => {
  if (jitter <= 5) return "text-emerald-600 dark:text-emerald-400";
  if (jitter <= 20) return "text-amber-600 dark:text-amber-500";
  if (jitter <= 50) return "text-orange-600 dark:text-orange-500";
  return "text-rose-600 dark:text-rose-500";
};

export function CustomNodeSelect({ 
  nodes, 
  value, 
  onChange 
}: { 
  nodes: NodeResult[]; 
  value: string; 
  onChange: (val: string) => void;
}) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const [alignment, setAlignment] = useState<"left" | "right">("left");
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    if (isOpen && dropdownRef.current) {
      const rect = dropdownRef.current.getBoundingClientRect();
      if (rect.right > window.innerWidth / 2) {
        setAlignment("right");
      } else {
        setAlignment("left");
      }
    }
  }, [isOpen]);

  const selectedNode = nodes.find(n => n.name === value) || nodes[0];

  return (
    <div className="relative w-full" ref={dropdownRef}>
      <div 
        onClick={() => setIsOpen(!isOpen)}
        className="group flex items-center justify-between w-full bg-background border border-border rounded-md px-3 py-2 text-sm cursor-pointer hover:border-amber-500/50 transition-colors shadow-sm"
      >
        <div className="flex items-center gap-2 flex-1 min-w-0 pr-2">
           {selectedNode?.provider && (
             <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md border border-border/50 shrink-0">
               {selectedNode.provider}
             </span>
           )}
           <span className="truncate font-medium text-foreground emoji-monochrome">{selectedNode ? selectedNode.name : t('ranking.select_node', 'Select node...')}</span>
        </div>
        <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
      </div>
      
      {isOpen && (
        <div className={`absolute top-full mt-1 min-w-[320px] max-h-60 overflow-y-auto overscroll-contain bg-background border border-border rounded-md shadow-lg z-50 animate-in fade-in zoom-in-95 duration-100 flex flex-col p-1 ${alignment === "right" ? "right-0" : "left-0"}`}>
          {nodes.map(node => (
            <div 
              key={node.name}
              onClick={() => { onChange(node.name); setIsOpen(false); }}
              className={`group flex items-center gap-2 px-2 py-2 cursor-pointer rounded-sm text-sm transition-colors hover:bg-muted ${node.name === value ? "bg-muted/50" : ""}`}
            >
              <div className="flex items-center gap-1.5 shrink-0">
                 {node.delay === null ? (
                   <span className="text-[11px] font-medium text-rose-500 dark:text-rose-400">
                     {node.backoff_rounds && node.backoff_rounds > 0 ? `Backoff: ${node.backoff_rounds}` : t('ranking.timeout', 'Timeout')}
                   </span>
                 ) : (
                   <>
                     <Zap className={`w-3.5 h-3.5 ${getColorClass(node.delay, "text")}`} />
                     <span className={`font-mono text-[13px] font-bold ${getColorClass(node.delay, "text")}`}>{node.delay}</span>
                   </>
                 )}
              </div>

              <div className="flex items-center gap-2 flex-1 min-w-0">
                {node.provider && (
                  <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md border border-border/50 shrink-0">
                    {node.provider}
                  </span>
                )}
                <span className="truncate text-foreground/90 emoji-monochrome">{node.name}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
