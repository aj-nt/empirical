const MODALITY_COLORS: Record<string, string> = {
  Cardinal: '#f85149',
  Fixed: '#3fb950',
  Mutable: '#58a6ff',
};

export function ModalityBar({ balance }: { balance: Record<string, number> }) {
  const total = Object.values(balance).reduce((a, b) => a + b, 0) || 1;
  return (
    <div className="space-y-2">
      {Object.entries(balance).map(([modality, count]) => (
        <div key={modality} className="flex items-center gap-2">
          <span className="text-xs w-16 text-muted">{modality}</span>
          <div className="flex-1 h-4 bg-surface rounded overflow-hidden">
            <div
              className="h-full rounded transition-all"
              style={{
                width: `${(count / total) * 100}%`,
                backgroundColor: MODALITY_COLORS[modality] || '#888',
              }}
            />
          </div>
          <span className="text-xs text-muted w-6 text-right">{count}</span>
        </div>
      ))}
    </div>
  );
}
