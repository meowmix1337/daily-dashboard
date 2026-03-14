import { useQuery } from '@tanstack/react-query';
import type { User } from '../types/auth';

async function fetchMe(): Promise<User> {
  const res = await fetch('/api/auth/me', { credentials: 'include' });
  if (!res.ok) throw new Error('unauthenticated');
  return res.json();
}

export function useAuth() {
  const { data: user, isLoading, isError } = useQuery<User>({
    queryKey: ['auth', 'me'],
    queryFn: fetchMe,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
  return { user: user ?? null, isLoading, isAuthenticated: !!user && !isError };
}
