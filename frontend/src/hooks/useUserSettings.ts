import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  fetchUserSettings,
  upsertUserSettings,
  fetchNewsCategories,
  setNewsCategories,
} from '../api/client';
import type { UserSettings, NewsCategoriesResponse } from '../types/dashboard';

export function useUserSettings() {
  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: fetchUserSettings,
  });
  return { settings, isLoading };
}

export function useNewsCategories() {
  const { data, isLoading } = useQuery({
    queryKey: ['news-categories'],
    queryFn: fetchNewsCategories,
  });
  return { data: data as NewsCategoriesResponse | undefined, isLoading };
}

export function useSettingsMutations() {
  const queryClient = useQueryClient();

  const save = useMutation({
    mutationFn: (body: Partial<UserSettings>) => upsertUserSettings(body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const saveCategories = useMutation({
    mutationFn: (categoryIds: string[]) => setNewsCategories(categoryIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['news-categories'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  return { save, saveCategories };
}
