import React from 'react';
import type { Quote } from '../types/dashboard';
import { Card } from './ui/Card';

interface QuoteCardProps {
  data: Quote;
  delay?: number;
}

export function QuoteCard({ data, delay = 0 }: QuoteCardProps): React.ReactElement {
  return (
    <Card delay={delay}>
      <div style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        padding: '8px 0',
      }}>
        <div style={{
          fontFamily: "'Playfair Display', Georgia, serif",
          fontSize: 20,
          fontStyle: 'italic',
          lineHeight: 1.6,
          color: '#d1d5db',
          marginBottom: 16,
        }}>
          "{data.text}"
        </div>
        <div style={{
          fontSize: 13,
          color: '#6b7280',
          fontWeight: 500,
        }}>
          — {data.author}
        </div>
      </div>
    </Card>
  );
}
