import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Dashboard from './components/Dashboard';
import { LoginPage } from './pages/LoginPage';
import { ProtectedRoute } from './components/ProtectedRoute';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  const isLoginPage = window.location.pathname === '/login';

  return (
    <QueryClientProvider client={queryClient}>
      {isLoginPage ? (
        <LoginPage />
      ) : (
        <ProtectedRoute>
          <Dashboard />
        </ProtectedRoute>
      )}
    </QueryClientProvider>
  );
}

export default App;
