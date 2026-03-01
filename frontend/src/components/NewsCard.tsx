import React, { useState } from 'react';
import { useNews } from '../hooks/useNews';
import { Card } from './ui/Card';
import { CardHeader } from './ui/CardHeader';

interface NewsCardProps {
  delay?: number;
}

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function Skeleton({ width = '100%', height = 16 }: { width?: string | number; height?: number }): React.ReactElement {
  return (
    <div style={{
      width,
      height,
      borderRadius: 4,
      background: 'rgba(255,255,255,0.07)',
      animation: 'pulse 1.5s ease-in-out infinite',
    }} />
  );
}

export function NewsCard({ delay = 0 }: NewsCardProps): React.ReactElement {
  const { data: categories, isLoading, isError, refetch, isFetching } = useNews();
  const [activeIndex, setActiveIndex] = useState(0);

  const active = categories?.[activeIndex] ?? { name: '', items: [] };

  return (
    <Card delay={delay} span={2}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <CardHeader icon="⊞" title="Headlines" />
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          style={{
            padding: '4px 10px',
            borderRadius: 6,
            border: '1px solid rgba(255,255,255,0.08)',
            background: 'rgba(255,255,255,0.04)',
            color: isFetching ? '#4b5563' : '#6b7280',
            fontSize: 12,
            cursor: isFetching ? 'not-allowed' : 'pointer',
            transition: 'all 0.15s',
          }}
        >
          {isFetching ? 'Refreshing…' : '↻ Refresh'}
        </button>
      </div>

      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} width={60} height={26} />
            ))}
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 8, padding: '14px 16px', borderRadius: 10, background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.05)' }}>
                <Skeleton width="30%" height={11} />
                <Skeleton width="100%" height={14} />
                <Skeleton width="80%" height={14} />
                <Skeleton width="40%" height={12} />
              </div>
            ))}
          </div>
        </div>
      ) : isError || !categories ? (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: 160, color: '#4b5563', fontSize: 13 }}>
          News unavailable
        </div>
      ) : (
        <>
          <div style={{
            display: 'flex',
            gap: 6,
            marginBottom: 16,
            overflowX: 'auto',
            scrollbarWidth: 'none',
            flexWrap: 'nowrap',
            paddingBottom: 2,
          }}>
            {categories.map((cat, i) => {
              const isActive = i === activeIndex;
              return (
                <button
                  key={cat.name}
                  onClick={() => setActiveIndex(i)}
                  style={{
                    flexShrink: 0,
                    padding: '5px 13px',
                    borderRadius: 20,
                    border: isActive ? '1px solid rgba(99,102,241,0.4)' : '1px solid rgba(255,255,255,0.08)',
                    background: isActive ? 'rgba(99,102,241,0.2)' : 'rgba(255,255,255,0.04)',
                    color: isActive ? '#818cf8' : '#6b7280',
                    fontSize: 12,
                    fontWeight: isActive ? 600 : 400,
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    letterSpacing: '0.02em',
                    lineHeight: 1,
                  }}
                >
                  {titleCase(cat.name)}
                </button>
              );
            })}
          </div>

          {active.items.length === 0 ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: 120, color: '#4b5563', fontSize: 13 }}>
              No articles available
            </div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              {active.items.map((item, i) => (
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
                  <div style={{ marginBottom: 8 }}>
                    <span style={{ fontSize: 11, color: '#4b5563' }}>{item.time}</span>
                  </div>
                  <div style={{ fontSize: 14, fontWeight: 500, color: '#e2e2e8', lineHeight: 1.45 }}>
                    {item.title}
                  </div>
                  <div style={{ fontSize: 12, color: '#6b7280', marginTop: 6 }}>
                    {item.source}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </Card>
  );
}
