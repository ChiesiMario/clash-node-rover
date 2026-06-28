import { useState, useEffect } from "react";
import { invoke } from "@tauri-apps/api/core";
import { useTranslation } from "react-i18next";
import { Loader2 } from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from "recharts";

interface NodeHistoryEntry {
  timestamp: string;
  node_name: string;
  delay: number | null;
  mean: number | null;
  jitter: number | null;
}

interface NodeHistoryChartProps {
  nodeName: string;
}

export function NodeHistoryChart({ nodeName }: NodeHistoryChartProps) {
  const { t } = useTranslation();
  const [data, setData] = useState<any[]>([]);
  const [hours, setHours] = useState<number>(24);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    invoke<NodeHistoryEntry[]>("get_node_history", { nodeName, hours })
      .then((history) => {
        if (!active) return;
        
        // Format data for recharts
        const formatted = history.map((entry) => {
          // Check if timestamp exists before creating Date object
          if (!entry.timestamp) return null;
          
          const date = new Date(entry.timestamp);
          
          // Recharts handles null values by breaking the line, which is what we want for Timeouts
          return {
            time: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            fullDate: date.toLocaleString(),
            score: entry.delay,
            mean: entry.mean,
            jitter: entry.jitter,
          };
        }).filter(Boolean); // Remove any null entries

        setData(formatted);
      })
      .catch(console.error)
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [nodeName, hours]);

  return (
    <div className="p-4 bg-muted/10 border-b border-border/50">
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-sm font-semibold">{t('ranking.history_chart', 'Historical Performance')}</h4>
        <div className="flex gap-1 bg-muted p-1 rounded-md border border-border/50">
          <button
            onClick={() => setHours(24)}
            className={`px-3 py-1 text-xs font-medium rounded-sm transition-colors ${hours === 24 ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"}`}
          >
            {t('ranking.24h', '24H')}
          </button>
          <button
            onClick={() => setHours(72)}
            className={`px-3 py-1 text-xs font-medium rounded-sm transition-colors ${hours === 72 ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"}`}
          >
            {t('ranking.3d', '3 Days')}
          </button>
          <button
            onClick={() => setHours(168)}
            className={`px-3 py-1 text-xs font-medium rounded-sm transition-colors ${hours === 168 ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"}`}
          >
            {t('ranking.7d', '7 Days')}
          </button>
        </div>
      </div>

      <div className="h-64 w-full relative">
        {loading ? (
          <div className="absolute inset-0 flex items-center justify-center">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
          </div>
        ) : data.length === 0 ? (
          <div className="absolute inset-0 flex items-center justify-center text-muted-foreground text-sm">
            {t('ranking.no_history', 'No historical data available yet.')}
          </div>
        ) : (
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="currentColor" className="text-border/40" vertical={false} />
              <XAxis 
                dataKey="time" 
                tick={{ fontSize: 11, fill: 'currentColor' }} 
                className="text-muted-foreground" 
                tickMargin={8}
                minTickGap={30}
              />
              <YAxis 
                tick={{ fontSize: 11, fill: 'currentColor' }} 
                className="text-muted-foreground"
                tickMargin={8}
                width={60}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: 'var(--background)', 
                  borderColor: 'var(--border)',
                  borderRadius: '8px',
                  boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'
                }}
                labelStyle={{ color: 'var(--foreground)', fontWeight: 600, marginBottom: '4px' }}
                itemStyle={{ fontSize: 13 }}
                labelFormatter={(value, payload) => {
                  if (payload && payload.length > 0 && payload[0].payload) {
                    return payload[0].payload.fullDate;
                  }
                  return value;
                }}
              />
              <Legend wrapperStyle={{ fontSize: 12, paddingTop: '10px' }} />
              <Line 
                type="monotone" 
                dataKey="score" 
                name={t('ranking.score', 'Score')} 
                stroke="#10b981" 
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 6 }}
                connectNulls={false}
              />
              <Line 
                type="monotone" 
                dataKey="mean" 
                name={t('ranking.mean', 'Mean')} 
                stroke="#f59e0b" 
                strokeWidth={2}
                dot={false}
                connectNulls={false}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  );
}
