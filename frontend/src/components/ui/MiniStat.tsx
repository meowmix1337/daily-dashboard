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
      background: 'rgba(255,255,255,0.03)',
      borderRadius: 8,
    }}>
      <div style={{ fontSize: 11, color: '#6b7280', marginBottom: 4 }}>{label}</div>
      <div style={{
        fontSize: 15,
        fontWeight: 600,
        color: accent || '#d1d5db',
        fontFamily: "'JetBrains Mono', monospace",
      }}>
        {value}
      </div>
    </div>
  );
}
