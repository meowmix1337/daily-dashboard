import React from 'react';
import type { NewsItem } from '../types/dashboard';
import { Card } from './ui/Card';
import { CardHeader } from './ui/CardHeader';

interface NewsCardProps {
  items: NewsItem[];
  delay?: number;
}

export function NewsCard({ items, delay = 0 }: NewsCardProps): React.ReactElement {
  return (
    <Card delay={delay} span={2}>
      <CardHeader icon="⊞" title="Headlines" />
      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: 12,
      }}>
        {items.map((item, i) => (
            <div
              key={i}
              onClick={() => item.url && window.open(item.url, '_blank', 'noopener,noreferrer')}
              style={{
                padding: '14px 16px',
                borderRadius: 10,
                background: 'rgba(255,255,255,0.02)',
                border: '1px solid rgba(255,255,255,0.05)',
                cursor: item.url ? 'pointer' : 'default',
                transition: 'all 0.2s',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                <span style={{ fontSize: 11, color: '#4b5563' }}>{item.time}</span>
              </div>
              <div style={{
                fontSize: 14,
                fontWeight: 500,
                color: '#e2e2e8',
                lineHeight: 1.45,
              }}>
                {item.title}
              </div>
              <div style={{ fontSize: 12, color: '#6b7280', marginTop: 6 }}>
                {item.source}
              </div>
            </div>
        ))}
      </div>
    </Card>
  );
}
