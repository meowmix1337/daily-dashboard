import React from 'react';

export function UnavailableCard({ span = 1, label }: { span?: number; label: string }): React.ReactElement {
  return (
    <div style={{
      gridColumn: `span ${span}`,
      background: 'rgba(255,255,255,0.025)',
      border: '1px solid rgba(255,255,255,0.06)',
      borderRadius: 16,
      padding: 24,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      gap: 8,
      minHeight: 120,
    }}>
      <span style={{ fontSize: 20, opacity: 0.3 }}>⚠</span>
      <div style={{ fontSize: 13, color: '#4b5563', textAlign: 'center' }}>{label}</div>
    </div>
  );
}
