import { useAuth } from '../hooks/useAuth';

interface Props {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: Props) {
  const { isLoading, isAuthenticated } = useAuth();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: '#0f172a', color: '#94a3b8' }}>
        Loading...
      </div>
    );
  }

  if (!isAuthenticated) {
    window.location.href = '/login';
    return null;
  }

  return <>{children}</>;
}
