import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { searchSymbols } from '../api/client';

export function useSymbolSearch() {
  const [rawQuery, setRawQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');

  useEffect(() => {
    if (!rawQuery.trim()) {
      setDebouncedQuery('');
      return;
    }
    const t = setTimeout(() => setDebouncedQuery(rawQuery.trim()), 350);
    return () => clearTimeout(t);
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
