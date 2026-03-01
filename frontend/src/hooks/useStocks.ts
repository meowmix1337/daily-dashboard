import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchStocks, addStockSymbol, removeStockSymbol } from '../api/client';
import type { StockQuote } from '../types/dashboard';

export function useStocks(initialStocks?: StockQuote[] | null) {
  const queryClient = useQueryClient();

  const query = useQuery<StockQuote[], Error>({
    queryKey: ['stocks'],
    queryFn: fetchStocks,
    initialData: initialStocks ?? undefined,
    staleTime: 0,
    refetchInterval: 10_000,
    retry: 2,
  });

  const add = useMutation({
    mutationFn: (symbol: string) => addStockSymbol(symbol),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stocks'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const remove = useMutation({
    mutationFn: (symbol: string) => removeStockSymbol(symbol),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stocks'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  return { stocks: query.data ?? [], isFetching: query.isFetching, add, remove };
}
