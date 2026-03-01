import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toggleTask, createTask, deleteTask } from '../api/client';
import type { DashboardResponse, Task } from '../types/dashboard';

export function useTasks() {
  const queryClient = useQueryClient();

  const toggle = useMutation({
    mutationFn: ({ id, done }: { id: string; done: boolean }) =>
      toggleTask(id, done),
    onMutate: async ({ id, done }) => {
      await queryClient.cancelQueries({ queryKey: ['dashboard'] });
      const previous = queryClient.getQueryData<DashboardResponse>(['dashboard']);

      queryClient.setQueryData<DashboardResponse>(['dashboard'], (old) => {
        if (!old) return old;
        return {
          ...old,
          tasks: old.tasks.map((t: Task) =>
            t.id === id ? { ...t, done } : t
          ),
        };
      });

      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['dashboard'], context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const create = useMutation({
    mutationFn: ({ text, priority }: { text: string; priority: string }) =>
      createTask(text, priority),
    onMutate: async ({ text, priority }) => {
      await queryClient.cancelQueries({ queryKey: ['dashboard'] });
      const previous = queryClient.getQueryData<DashboardResponse>(['dashboard']);

      const tempTask: Task = { id: `temp-${Date.now()}`, text, done: false, priority: priority as Task['priority'] };
      queryClient.setQueryData<DashboardResponse>(['dashboard'], (old) => {
        if (!old) return old;
        return { ...old, tasks: [...old.tasks, tempTask] };
      });

      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['dashboard'], context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) => deleteTask(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['dashboard'] });
      const previous = queryClient.getQueryData<DashboardResponse>(['dashboard']);

      queryClient.setQueryData<DashboardResponse>(['dashboard'], (old) => {
        if (!old) return old;
        return { ...old, tasks: old.tasks.filter((t: Task) => t.id !== id) };
      });

      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['dashboard'], context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  return { toggle, create, remove };
}
