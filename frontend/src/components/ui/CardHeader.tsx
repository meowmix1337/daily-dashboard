import React from 'react';

interface CardHeaderProps {
  icon: string;
  title: string;
  badge?: string;
}

export function CardHeader({ icon, title, badge }: CardHeaderProps): React.ReactElement {
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginBottom: 16,
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{ fontSize: 14, color: 'var(--text-accent)', opacity: 0.8 }}>{icon}</span>
        <span style={{
          fontSize: 13,
          fontWeight: 600,
          color: 'var(--text-secondary)',
          letterSpacing: '0.04em',
          textTransform: 'uppercase',
        }}>
          {title}
        </span>
      </div>
      {badge && (
        <span style={{
          fontSize: 11,
          color: 'var(--text-secondary)',
          fontFamily: "'JetBrains Mono', monospace",
          fontWeight: 400,
        }}>
          {badge}
        </span>
      )}
    </div>
  );
}
