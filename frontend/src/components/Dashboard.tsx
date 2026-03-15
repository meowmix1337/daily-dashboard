import React, { useEffect, useState } from 'react';
import { useDashboard } from '../hooks/useDashboard';
import { useClock } from '../hooks/useClock';
import { useAuth } from '../hooks/useAuth';
import { WeatherCard } from './WeatherCard';
import { CalendarCard } from './CalendarCard';
import { TasksCard } from './TasksCard';
import { NewsCard } from './NewsCard';
import { StocksCard } from './StocksCard';
import { QuoteCard } from './QuoteCard';
import { UserProfile } from './UserProfile';
import { UnavailableCard } from './ui/UnavailableCard';

function getGreeting(date: Date): string {
  const h = date.getHours();
  if (h < 12) return 'Good morning';
  if (h < 17) return 'Good afternoon';
  return 'Good evening';
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: true,
  });
}

function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });
}

// Skeleton pulse component
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

// Card skeleton for loading state
function CardSkeleton({ span = 1, rows = 3 }: { span?: number; rows?: number }): React.ReactElement {
  return (
    <div style={{
      gridColumn: `span ${span}`,
      background: 'rgba(255,255,255,0.025)',
      border: '1px solid rgba(255,255,255,0.06)',
      borderRadius: 16,
      padding: 24,
    }}>
      <Skeleton height={14} width="40%" />
      <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
        {Array.from({ length: rows }).map((_, i) => (
          <Skeleton key={i} height={14} width={`${70 + (i % 3) * 10}%`} />
        ))}
      </div>
    </div>
  );
}

export default function Dashboard(): React.ReactElement {
  const { data, isLoading, isError, refetch } = useDashboard();
  const { user } = useAuth();
  const now = useClock();
  const [headerLoaded, setHeaderLoaded] = useState(false);

  useEffect(() => {
    const t = setTimeout(() => setHeaderLoaded(true), 100);
    return () => clearTimeout(t);
  }, []);

  const lastUpdated = formatTime(now);

  return (
    <div style={{
      minHeight: '100vh',
      background: '#0a0a0f',
      color: '#e2e2e8',
      fontFamily: "'DM Sans', 'Helvetica Neue', sans-serif",
      padding: 32,
      position: 'relative',
    }}>
      {/* Ambient background gradients */}
      <div style={{
        position: 'fixed',
        top: 0, left: 0, right: 0, bottom: 0,
        pointerEvents: 'none',
        zIndex: 0,
        background: `
          radial-gradient(ellipse 600px 400px at 15% 20%, rgba(99,102,241,0.07) 0%, transparent 70%),
          radial-gradient(ellipse 500px 500px at 85% 70%, rgba(236,72,153,0.05) 0%, transparent 70%),
          radial-gradient(ellipse 400px 300px at 50% 90%, rgba(16,185,129,0.04) 0%, transparent 70%)
        `,
      }} />

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
      `}</style>

      <div style={{ position: 'relative', zIndex: 1, maxWidth: 1440, margin: '0 auto' }}>
        {/* Header */}
        <div style={{
          opacity: headerLoaded ? 1 : 0,
          transform: headerLoaded ? 'translateY(0)' : 'translateY(12px)',
          transition: 'all 0.8s cubic-bezier(0.16, 1, 0.3, 1)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-end',
          marginBottom: 40,
          borderBottom: '1px solid rgba(255,255,255,0.06)',
          paddingBottom: 24,
          position: 'relative',
          zIndex: 10,
        }}>
          <div>
            <div style={{
              fontSize: 14,
              fontWeight: 500,
              color: '#6366f1',
              letterSpacing: '0.12em',
              textTransform: 'uppercase',
              marginBottom: 8,
            }}>
              {formatDate(now)}
            </div>
            <h1 style={{
              fontFamily: "'Playfair Display', Georgia, serif",
              fontSize: 42,
              fontWeight: 700,
              margin: 0,
              background: 'linear-gradient(135deg, #e2e2e8 0%, #9ca3af 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}>
              {getGreeting(now)}
            </h1>
          </div>
          <div style={{ display: 'flex', alignItems: 'flex-end', gap: 16 }}>
            <div style={{ textAlign: 'right' }}>
              <div style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 32,
                fontWeight: 500,
                color: '#f0f0f5',
                letterSpacing: '-0.02em',
              }}>
                {formatTime(now)}
              </div>
              <div style={{ fontSize: 13, color: '#6b7280', marginTop: 4 }}>
                {data?.meta?.sunrise
                  ? `☀️ ${data.meta.sunrise} → 🌙 ${data.meta.sunset} · ${data.meta.daylight} daylight`
                  : '☀️ — → 🌙 — · — daylight'}
              </div>
            </div>
            {user && <UserProfile user={user} />}
          </div>
        </div>

        {/* Error banner */}
        {isError && (
          <div style={{
            marginBottom: 20,
            padding: '12px 16px',
            borderRadius: 10,
            background: 'rgba(239,68,68,0.1)',
            border: '1px solid rgba(239,68,68,0.2)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}>
            <span style={{ fontSize: 14, color: '#ef4444' }}>
              Failed to load dashboard data
            </span>
            <button
              onClick={() => refetch()}
              style={{
                fontSize: 13,
                color: '#ef4444',
                background: 'rgba(239,68,68,0.15)',
                border: '1px solid rgba(239,68,68,0.3)',
                borderRadius: 6,
                padding: '4px 12px',
                cursor: 'pointer',
              }}
            >
              Retry
            </button>
          </div>
        )}

        {/* Ticker bar */}
        <StocksCard stocks={data?.stocks ?? null} delay={0.1} />

        {/* Grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr 1fr',
          gap: 20,
        }}>
          {isLoading ? (
            <>
              <CardSkeleton span={1} rows={4} />
              <CardSkeleton span={1} rows={5} />
              <CardSkeleton span={1} rows={5} />
              <CardSkeleton span={2} rows={4} />
              <CardSkeleton span={1} rows={2} />
            </>
          ) : (
            <>
              {/* Row 1 */}
              {data?.weather ? (
                <WeatherCard data={data.weather} delay={0.2} />
              ) : (
                <UnavailableCard span={1} label="Weather unavailable" />
              )}
              {data?.calendar ? (
                <CalendarCard events={data.calendar} delay={0.3} />
              ) : (
                <CardSkeleton span={1} rows={5} />
              )}
              {data?.tasks ? (
                <TasksCard tasks={data.tasks} delay={0.4} />
              ) : (
                <CardSkeleton span={1} rows={5} />
              )}

              {/* Row 2 */}
              <NewsCard delay={0.5} />
              {data?.meta?.quote?.text ? (
                <QuoteCard data={data.meta.quote} delay={0.6} />
              ) : (
                <UnavailableCard span={1} label="Quote unavailable" />
              )}
            </>
          )}
        </div>

        {/* Footer */}
        <div style={{
          marginTop: 32,
          paddingTop: 20,
          borderTop: '1px solid rgba(255,255,255,0.06)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          opacity: 0.5,
        }}>
          <div style={{ fontSize: 12, color: '#4b5563' }}>
            Daily Dashboard · Powered by Open-Meteo, GNews, Finnhub, Google APIs
          </div>
          <div style={{ fontSize: 12, color: '#4b5563' }}>
            Last updated: {lastUpdated}
          </div>
        </div>
      </div>
    </div>
  );
}
