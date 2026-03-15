import React from 'react';

export function UnavailableCard({ span = 1, label }: { span?: number; label: string }): React.ReactElement {
  return (
    <div style={{
      gridColumn: `span ${span}`,
      background: 'var(--bg-card)',
      border: '1px solid var(--bg-card-border)',
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
      <div style={{ fontSize: 13, color: 'var(--text-muted)', textAlign: 'center' }}>{label}</div>
    </div>
  );
}
