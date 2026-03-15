import React from 'react';
import type { WeatherData } from '../types/dashboard';
import { Card } from './ui/Card';
import { CardHeader } from './ui/CardHeader';
import { MiniStat } from './ui/MiniStat';

interface WeatherCardProps {
  data: WeatherData;
  delay?: number;
}

export function WeatherCard({ data, delay = 0 }: WeatherCardProps): React.ReactElement {
  const aqiColor = data.aqi <= 50 ? '#10b981' : data.aqi <= 100 ? '#f59e0b' : '#ef4444';

  return (
    <Card delay={delay}>
      <CardHeader icon="◐" title="Weather" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 20 }}>
        <span style={{ fontSize: 56, lineHeight: 1 }}>{data.icon}</span>
        <div>
          <div style={{
            fontSize: 48,
            fontWeight: 300,
            fontFamily: "'DM Sans', sans-serif",
            lineHeight: 1,
            color: 'var(--text-clock)',
          }}>
            {Math.round(data.temp)}°
          </div>
          <div style={{ color: 'var(--text-secondary)', fontSize: 14, marginTop: 4 }}>
            {data.condition} · H:{Math.round(data.high)}° L:{Math.round(data.low)}°
          </div>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <MiniStat label="Humidity" value={`${data.humidity}%`} />
        <MiniStat label="Wind" value={`${data.windSpeed} mph`} />
        <MiniStat label="UV" value={data.uvIndex} />
        <MiniStat label="AQI" value={data.aqi} accent={aqiColor} />
      </div>
      <div style={{
        display: 'flex',
        gap: 0,
        borderTop: '1px solid var(--border-subtle)',
        paddingTop: 12,
      }}>
        {data.hourly.slice(0, 8).map((h, i) => (
          <div key={i} style={{ flex: 1, textAlign: 'center', padding: '4px 0' }}>
            <div style={{ fontSize: 11, color: 'var(--text-secondary)' }}>{h.time}</div>
            <div style={{ fontSize: 16, margin: '4px 0' }}>{h.icon}</div>
            <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--text-primary)' }}>
              {Math.round(h.temp)}°
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}
