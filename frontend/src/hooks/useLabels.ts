import { useMemo } from 'react';
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  assignLabelToTask,
  createLabel as apiCreateLabel,
  deleteLabel as apiDeleteLabel,
  fetchLabels,
  fetchTaskLabels,
  removeLabelFromTask,
} from '../api/client';
import type { TaskLabel } from '../types/dashboard';

export function useLabels() {
  const { data, isLoading } = useQuery<TaskLabel[]>({
    queryKey: ['labels'],
    queryFn: fetchLabels,
  });
  return { labels: data ?? [], isLoading };
}

export function useTaskLabels(taskIds: string[]): Map<string, TaskLabel[]> {
  const results = useQueries({
    queries: taskIds.map((id) => ({
      queryKey: ['task-labels', id],
      queryFn: () => fetchTaskLabels(id),
      staleTime: 30_000,
    })),
  });

  return useMemo(() => {
    const map = new Map<string, TaskLabel[]>();
    taskIds.forEach((id, i) => {
      const result = results[i];
      map.set(id, result?.data ?? []);
    });
    return map;
  }, [taskIds, results]);
}

export function useLabelMutations() {
  const queryClient = useQueryClient();

  const createLabel = useMutation({
    mutationFn: ({ name, color }: { name: string; color: string }) =>
      apiCreateLabel(name, color),
    onMutate: async ({ name, color }) => {
      await queryClient.cancelQueries({ queryKey: ['labels'] });
      const previous = queryClient.getQueryData<TaskLabel[]>(['labels']);
      const tempLabel: TaskLabel = {
        id: `temp-${Date.now()}`,
        name,
        color,
        created_at: new Date().toISOString(),
      };
      queryClient.setQueryData<TaskLabel[]>(['labels'], (old) =>
        old ? [...old, tempLabel] : [tempLabel]
      );
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) queryClient.setQueryData(['labels'], ctx.previous);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['labels'] });
    },
  });

  const deleteLabel = useMutation({
    mutationFn: (id: string) => apiDeleteLabel(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['labels'] });
      const previous = queryClient.getQueryData<TaskLabel[]>(['labels']);
      queryClient.setQueryData<TaskLabel[]>(['labels'], (old) =>
        old ? old.filter((l) => l.id !== id) : []
      );
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) queryClient.setQueryData(['labels'], ctx.previous);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['labels'] });
      queryClient.invalidateQueries({ queryKey: ['task-labels'] });
    },
  });

  const assignLabel = useMutation({
    mutationFn: ({ taskId, labelId }: { taskId: string; labelId: string }) =>
      assignLabelToTask(taskId, labelId),
    onSettled: (_data, _err, { taskId }) => {
      queryClient.invalidateQueries({ queryKey: ['task-labels', taskId] });
    },
  });

  const removeLabel = useMutation({
    mutationFn: ({ taskId, labelId }: { taskId: string; labelId: string }) =>
      removeLabelFromTask(taskId, labelId),
    onMutate: async ({ taskId, labelId }) => {
      await queryClient.cancelQueries({ queryKey: ['task-labels', taskId] });
      const previous = queryClient.getQueryData<TaskLabel[]>(['task-labels', taskId]);
      queryClient.setQueryData<TaskLabel[]>(['task-labels', taskId], (old) =>
        old ? old.filter((l) => l.id !== labelId) : []
      );
      return { previous };
    },
    onError: (_err, { taskId }, ctx) => {
      if (ctx?.previous)
        queryClient.setQueryData(['task-labels', taskId], ctx.previous);
    },
    onSettled: (_data, _err, { taskId }) => {
      queryClient.invalidateQueries({ queryKey: ['task-labels', taskId] });
    },
  });

  return { createLabel, deleteLabel, assignLabel, removeLabel };
}
