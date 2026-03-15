import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function LoginPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);

  // Show nothing while auth check is in flight to prevent flash of login UI
  if (isLoading) return null;

  // Already authenticated — redirect to dashboard immediately (no useEffect delay)
  if (isAuthenticated) return <Navigate to="/" replace />;

  const cardStyle = {
    background: 'var(--bg-card)',
    backdropFilter: 'blur(20px)',
    border: '1px solid var(--bg-card-border)',
    borderRadius: 16,
    padding: 48,
    textAlign: 'center',
    maxWidth: 400,
    width: '100%',
  } satisfies React.CSSProperties;

  const buttonStyle = {
    display: 'flex',
    alignItems: 'center',
    gap: 12,
    background: hovered ? '#f1f5f9' : '#ffffff',
    color: '#374151',
    border: 'none',
    borderRadius: 8,
    padding: '12px 24px',
    fontSize: '1rem',
    fontWeight: 600,
    cursor: 'pointer',
    width: '100%',
    justifyContent: 'center',
    transition: 'background 0.15s ease, transform 0.1s ease, outline-color 0.1s ease',
    outline: focused ? '2px solid #3b82f6' : '2px solid transparent',
    outlineOffset: 2,
  } satisfies React.CSSProperties;

  return (
    <div style={{
      minHeight: '100vh',
      background: 'var(--bg-primary)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '0 16px',
    }}>
      <div style={cardStyle}>
        <h1 style={{ color: 'var(--text-primary)', fontSize: '1.5rem', fontWeight: 700, marginBottom: 8, marginTop: 0 }}>
          Daily Dashboard
        </h1>
        <p style={{ color: 'var(--text-secondary)', marginBottom: 32, marginTop: 0, fontSize: '1rem' }}>
          Sign in to access your personal dashboard
        </p>
        <button
          type="button"
          onClick={() => { window.location.href = '/api/auth/login'; }}
          onMouseEnter={() => setHovered(true)}
          onMouseLeave={() => setHovered(false)}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          style={buttonStyle}
        >
          <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
            <path fill="#4285F4" d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.874 2.684-6.615z"/>
            <path fill="#34A853" d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.258c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332C2.438 15.983 5.482 18 9 18z"/>
            <path fill="#FBBC05" d="M3.964 10.707c-.18-.54-.282-1.117-.282-1.707s.102-1.167.282-1.707V4.961H.957C.347 6.175 0 7.55 0 9s.348 2.825.957 4.039l3.007-2.332z"/>
            <path fill="#EA4335" d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0 5.482 0 2.438 2.017.957 4.961L3.964 7.293C4.672 5.166 6.656 3.58 9 3.58z"/>
          </svg>
          Sign in with Google
        </button>
      </div>
    </div>
  );
}
