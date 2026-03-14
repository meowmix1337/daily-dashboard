import { useEffect } from 'react';
import { useAuth } from '../hooks/useAuth';

interface Props {
  children: React.ReactNode;
}

// Spinner keyframes injected once at module level
const spinnerStyle = document.createElement('style');
spinnerStyle.textContent = '@keyframes spin { to { transform: rotate(360deg); } }';
if (!document.head.querySelector('[data-spinner]')) {
  spinnerStyle.setAttribute('data-spinner', '');
  document.head.appendChild(spinnerStyle);
}

export function ProtectedRoute({ children }: Props) {
  const { isLoading, isAuthenticated } = useAuth();

  // All navigation uses full-reload (no react-router); window.location.href is the app convention.
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      window.location.href = '/login';
    }
  }, [isLoading, isAuthenticated]);

  if (isLoading) {
    return (
      <div
        role="status"
        aria-label="Loading"
        aria-live="polite"
        style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: '#0f172a' }}
      >
        <div style={{
          width: 32,
          height: 32,
          borderRadius: '50%',
          border: '3px solid rgba(255,255,255,0.1)',
          borderTopColor: '#94a3b8',
          animation: 'spin 0.8s linear infinite',
        }} />
      </div>
    );
  }

  // Not authenticated — return null while the useEffect fires the redirect
  if (!isAuthenticated) return null;

  return <>{children}</>;
}
