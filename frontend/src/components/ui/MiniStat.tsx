import React from 'react';

interface MiniStatProps {
  label: string;
  value: string | number;
  accent?: string;
}

export function MiniStat({ label, value, accent }: MiniStatProps): React.ReactElement {
  return (
    <div style={{
      flex: 1,
      textAlign: 'center',
      padding: '8px 4px',
      background: 'var(--bg-card)',
      borderRadius: 8,
    }}>
      <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginBottom: 4 }}>{label}</div>
      <div style={{
        fontSize: 15,
        fontWeight: 600,
        color: accent || 'var(--text-primary)',
        fontFamily: "'JetBrains Mono', monospace",
      }}>
        {value}
      </div>
    </div>
  );
}
