import React from 'react';
import type { CalendarEvent } from '../types/dashboard';
import { Card } from './ui/Card';
import { CardHeader } from './ui/CardHeader';

interface CalendarCardProps {
  events: CalendarEvent[];
  delay?: number;
}

export function CalendarCard({ events, delay = 0 }: CalendarCardProps): React.ReactElement {
  return (
    <Card delay={delay}>
      <CardHeader icon="▦" title="Today's Schedule" badge={`${events.length} events`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {events.map((event, i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '10px 12px',
              borderRadius: 10,
              background: i === 0 ? 'rgba(99,102,241,0.1)' : 'rgba(255,255,255,0.02)',
              border: i === 0
                ? '1px solid rgba(99,102,241,0.2)'
                : '1px solid transparent',
              transition: 'all 0.2s',
            }}
          >
            <div style={{
              width: 3,
              height: 32,
              borderRadius: 2,
              background: event.color,
              flexShrink: 0,
            }} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{
                fontSize: 14,
                fontWeight: 500,
                color: '#e2e2e8',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}>
                {event.title}
              </div>
              <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>
                {event.time} · {event.duration}
              </div>
            </div>
            {i === 0 && (
              <div style={{
                fontSize: 10,
                fontWeight: 600,
                color: '#6366f1',
                background: 'rgba(99,102,241,0.15)',
                padding: '3px 8px',
                borderRadius: 20,
                letterSpacing: '0.05em',
                textTransform: 'uppercase',
                flexShrink: 0,
              }}>
                Next
              </div>
            )}
          </div>
        ))}
      </div>
    </Card>
  );
}
