import { useQuery } from '@tanstack/react-query';
import type { User } from '../types/auth';

export class AuthError extends Error {
  status: number;
  constructor(status: number) {
    super('unauthenticated');
    this.name = 'AuthError';
    this.status = status;
  }
}

async function fetchMe(): Promise<User> {
  const res = await fetch('/api/auth/me', { credentials: 'include' });
  if (res.status === 401) throw new AuthError(401);
  if (!res.ok) throw new Error(`unexpected status ${res.status}`);
  return res.json();
}

export function useAuth() {
  const { data: user, isLoading, error } = useQuery<User, Error>({
    queryKey: ['auth', 'me'],
    queryFn: fetchMe,
    retry: false,
    staleTime: 30_000,
    refetchOnWindowFocus: true,
  });
  const isAuthenticated = !!user && !(error instanceof AuthError);
  return { user: user ?? null, isLoading, isAuthenticated, error };
}
