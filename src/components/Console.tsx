

export function Console() {
  return (
    <div className="h-full flex flex-col p-8 max-w-5xl mx-auto space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">System Logs</h1>
        <p className="text-muted-foreground">Detailed execution traces from the background engine.</p>
      </div>

      <div className="flex-1 bg-[#0c0c0c] border border-border rounded-xl shadow-inner p-4 font-mono text-sm text-gray-300 overflow-auto flex flex-col gap-2">
        <div className="flex gap-4 opacity-50">
          <span className="text-blue-400">00:00:00</span>
          <span className="text-emerald-400">[INFO]</span>
          <span>Node Rover Watchdog initialized in Rust.</span>
        </div>
        <div className="flex gap-4">
          <span className="text-blue-400">00:00:01</span>
          <span className="text-amber-400">[WARN]</span>
          <span>Waiting for valid API configuration...</span>
        </div>
      </div>
    </div>
  );
}
