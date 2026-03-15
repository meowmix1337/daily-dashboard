import { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { searchSymbols } from '../api/client';

export function useSymbolSearch() {
  const [rawQuery, setRawQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (timerRef.current !== null) clearTimeout(timerRef.current);
    const trimmed = rawQuery.trim();
    timerRef.current = setTimeout(() => {
      setDebouncedQuery(trimmed);
    }, trimmed ? 350 : 0);
    return () => {
      if (timerRef.current !== null) clearTimeout(timerRef.current);
    };
  }, [rawQuery]);

  const { data, isFetching } = useQuery({
    queryKey: ['stocks', 'search', debouncedQuery],
    queryFn: () => searchSymbols(debouncedQuery),
    enabled: debouncedQuery.length > 0,
    staleTime: 30_000,
    retry: 1,
  });

  return {
    searchQuery: rawQuery,
    setSearchQuery: setRawQuery,
    results: data?.results ?? [],
    isSearching: isFetching,
  };
}
