import React, { useEffect, useState } from 'react';
import { useDashboard } from '../hooks/useDashboard';
import { useClock } from '../hooks/useClock';
import { useAuth } from '../hooks/useAuth';
import { useTheme } from '../hooks/useTheme';
import { useWindowSize } from '../hooks/useWindowSize';
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
      background: 'var(--bg-skeleton)',
      animation: 'pulse 1.5s ease-in-out infinite',
    }} />
  );
}

// Card skeleton for loading state
function CardSkeleton({ span = 1, rows = 3 }: { span?: number; rows?: number }): React.ReactElement {
  return (
    <div style={{
      gridColumn: `span ${span}`,
      background: 'var(--bg-card)',
      border: '1px solid var(--bg-card-border)',
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
  const { theme, toggleTheme } = useTheme();
  const now = useClock();
  const { breakpoint } = useWindowSize();
  const [headerLoaded, setHeaderLoaded] = useState(false);
  const [toggleHovered, setToggleHovered] = useState(false);

  const isMobile = breakpoint === 'mobile';
  const isTablet = breakpoint === 'tablet';

  useEffect(() => {
    const t = setTimeout(() => setHeaderLoaded(true), 100);
    return () => clearTimeout(t);
  }, []);

  const lastUpdated = formatTime(now);

  return (
    <div style={{
      minHeight: '100vh',
      background: 'var(--bg-primary)',
      color: 'var(--text-primary)',
      fontFamily: "'DM Sans', 'Helvetica Neue', sans-serif",
      padding: isMobile ? 16 : isTablet ? 24 : 32,
      position: 'relative',
    }}>
      {/* Ambient background gradients */}
      <div style={{
        position: 'fixed',
        top: 0, left: 0, right: 0, bottom: 0,
        pointerEvents: 'none',
        zIndex: 0,
        background: `
          radial-gradient(ellipse 600px 400px at 15% 20%, var(--ambient-indigo) 0%, transparent 70%),
          radial-gradient(ellipse 500px 500px at 85% 70%, var(--ambient-pink) 0%, transparent 70%),
          radial-gradient(ellipse 400px 300px at 50% 90%, var(--ambient-green) 0%, transparent 70%)
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
          flexDirection: isMobile ? 'column' : 'row',
          justifyContent: 'space-between',
          alignItems: isMobile ? 'flex-start' : 'flex-end',
          gap: isMobile ? 16 : 0,
          marginBottom: isMobile ? 24 : 40,
          borderBottom: '1px solid var(--header-border)',
          paddingBottom: isMobile ? 16 : 24,
          position: 'relative',
          zIndex: 10,
        }}>
          <div>
            <div style={{
              fontSize: isMobile ? 12 : 14,
              fontWeight: 500,
              color: 'var(--text-accent)',
              letterSpacing: '0.12em',
              textTransform: 'uppercase',
              marginBottom: 8,
            }}>
              {formatDate(now)}
            </div>
            <h1 style={{
              fontFamily: "'Playfair Display', Georgia, serif",
              fontSize: isMobile ? 28 : isTablet ? 34 : 42,
              fontWeight: 700,
              margin: 0,
              background: 'linear-gradient(135deg, var(--gradient-heading-start) 0%, var(--gradient-heading-end) 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
            }}>
              {getGreeting(now)}
            </h1>
          </div>
          <div style={{ display: 'flex', alignItems: isMobile ? 'center' : 'flex-end', gap: 16, alignSelf: isMobile ? 'stretch' : 'auto', justifyContent: isMobile ? 'space-between' : 'flex-end' }}>
            <div style={{ textAlign: isMobile ? 'left' : 'right' }}>
              <div style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: isMobile ? 22 : isTablet ? 26 : 32,
                fontWeight: 500,
                color: 'var(--text-clock)',
                letterSpacing: '-0.02em',
              }}>
                {formatTime(now)}
              </div>
              <div style={{ fontSize: isMobile ? 11 : 13, color: 'var(--text-secondary)', marginTop: 4 }}>
                {data?.meta?.sunrise
                  ? `☀️ ${data.meta.sunrise} → 🌙 ${data.meta.sunset} · ${data.meta.daylight} daylight`
                  : '☀️ — → 🌙 — · — daylight'}
              </div>
            </div>
            <button
              onClick={toggleTheme}
              aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              onMouseEnter={() => setToggleHovered(true)}
              onMouseLeave={() => setToggleHovered(false)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: 36,
                height: 36,
                borderRadius: 8,
                background: toggleHovered ? 'var(--toggle-hover-bg)' : 'var(--toggle-bg)',
                border: '1px solid var(--toggle-border)',
                color: 'var(--toggle-text)',
                cursor: 'pointer',
                fontSize: 16,
                flexShrink: 0,
                transition: 'background 0.2s ease',
              }}
            >
              {theme === 'dark' ? '☀️' : '🌙'}
            </button>
            {user && <UserProfile user={user} />}
          </div>
        </div>

        {/* Error banner */}
        {isError && (
          <div style={{
            marginBottom: 20,
            padding: '12px 16px',
            borderRadius: 10,
            background: 'var(--error-bg)',
            border: '1px solid var(--error-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}>
            <span style={{ fontSize: 14, color: 'var(--error-text)' }}>
              Failed to load dashboard data
            </span>
            <button
              onClick={() => refetch()}
              style={{
                fontSize: 13,
                color: 'var(--error-text)',
                background: 'var(--error-button-bg)',
                border: '1px solid var(--error-button-border)',
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
          gridTemplateColumns: isMobile ? '1fr' : isTablet ? '1fr 1fr' : '1fr 1fr 1fr',
          gap: isMobile ? 12 : 20,
        }}>
          {isLoading ? (
            <>
              <CardSkeleton span={1} rows={4} />
              <CardSkeleton span={1} rows={5} />
              <CardSkeleton span={1} rows={5} />
              <CardSkeleton span={isMobile || isTablet ? 1 : 2} rows={4} />
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
              <NewsCard delay={0.5} isMobile={isMobile} isTablet={isTablet} />
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
          marginTop: isMobile ? 16 : 32,
          paddingTop: 20,
          borderTop: '1px solid var(--footer-border)',
          display: 'flex',
          flexDirection: isMobile ? 'column' : 'row',
          justifyContent: 'space-between',
          alignItems: isMobile ? 'flex-start' : 'center',
          gap: isMobile ? 4 : 0,
          opacity: 0.5,
        }}>
          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
            Daily Dashboard · Powered by Open-Meteo, GNews, Finnhub, Google APIs
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
            Last updated: {lastUpdated}
          </div>
        </div>
      </div>
    </div>
  );
}
