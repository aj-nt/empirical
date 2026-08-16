const ELEMENT_COLORS: Record<string, string> = {
  Fire: '#f85149',
  Earth: '#3fb950',
  Air: '#f0c040',
  Water: '#58a6ff',
};

export function ElementBar({ balance }: { balance: Record<string, number> }) {
  const total = Object.values(balance).reduce((a, b) => a + b, 0) || 1;
  return (
    <div className="space-y-2">
      {Object.entries(balance).map(([element, count]) => (
        <div key={element} className="flex items-center gap-2">
          <span className="text-xs w-12 text-muted">{element}</span>
          <div className="flex-1 h-4 bg-surface rounded overflow-hidden">
            <div
              className="h-full rounded transition-all"
              style={{
                width: `${(count / total) * 100}%`,
                backgroundColor: ELEMENT_COLORS[element] || '#888',
              }}
            />
          </div>
          <span className="text-xs text-muted w-6 text-right">{count}</span>
        </div>
      ))}
    </div>
  );
}
