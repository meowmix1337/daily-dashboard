import React, { useState } from 'react';
import type { Task } from '../types/dashboard';
import { Card } from './ui/Card';
import { CardHeader } from './ui/CardHeader';
import { useTasks } from '../hooks/useTasks';

interface TasksCardProps {
  tasks: Task[];
  delay?: number;
}

const PRIORITY_COLOR: Record<string, string> = {
  high: '#ef4444',
  medium: '#f59e0b',
  low: '#6b7280',
};

const PRIORITY_CYCLE: Task['priority'][] = ['high', 'medium', 'low'];

export function TasksCard({ tasks, delay = 0 }: TasksCardProps): React.ReactElement {
  const { toggle, create, remove } = useTasks();
  const [newText, setNewText] = useState('');
  const [newPriority, setNewPriority] = useState<Task['priority']>('medium');
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  const doneCount = tasks.filter((t) => t.done).length;

  function cyclePriority() {
    const idx = PRIORITY_CYCLE.indexOf(newPriority);
    setNewPriority(PRIORITY_CYCLE[(idx + 1) % PRIORITY_CYCLE.length]);
  }

  function handleAdd() {
    const text = newText.trim();
    if (!text) return;
    create.mutate({ text, priority: newPriority });
    setNewText('');
  }

  return (
    <Card delay={delay}>
      <CardHeader
        icon="◉"
        title="Tasks"
        badge={`${doneCount}/${tasks.length} done`}
      />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {tasks.map((task) => (
          <div
            key={task.id}
            onMouseEnter={() => setHoveredId(task.id)}
            onMouseLeave={() => setHoveredId(null)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '10px 12px',
              borderRadius: 10,
              background: task.done
                ? 'rgba(16,185,129,0.05)'
                : 'var(--bg-card)',
              transition: 'all 0.2s',
            }}
          >
            {/* Checkbox */}
            <div
              onClick={() => toggle.mutate({ id: task.id, done: !task.done })}
              style={{
                width: 20,
                height: 20,
                borderRadius: 6,
                flexShrink: 0,
                border: task.done ? '2px solid #10b981' : '2px solid var(--border-medium)',
                background: task.done ? '#10b981' : 'transparent',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                transition: 'all 0.2s',
                fontSize: 12,
                color: 'var(--bg-primary)',
              }}
            >
              {task.done && '✓'}
            </div>

            {/* Text */}
            <div
              onClick={() => toggle.mutate({ id: task.id, done: !task.done })}
              style={{
                flex: 1,
                fontSize: 13,
                color: task.done ? 'var(--text-muted)' : 'var(--text-primary)',
                textDecoration: task.done ? 'line-through' : 'none',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {task.text}
            </div>

            {/* Priority dot */}
            <div style={{
              width: 7,
              height: 7,
              borderRadius: '50%',
              flexShrink: 0,
              background: PRIORITY_COLOR[task.priority] ?? 'var(--text-secondary)',
              opacity: task.done ? 0.3 : 1,
            }} />

            {/* Delete button */}
            <button
              onClick={() => remove.mutate(task.id)}
              style={{
                background: 'none',
                border: 'none',
                padding: '0 2px',
                cursor: 'pointer',
                fontSize: 14,
                lineHeight: 1,
                color: 'var(--text-secondary)',
                opacity: hoveredId === task.id ? 0.7 : 0,
                transition: 'opacity 0.15s',
                flexShrink: 0,
              }}
              title="Delete task"
            >
              ×
            </button>
          </div>
        ))}
      </div>

      {/* Add task form */}
      <div style={{
        marginTop: 12,
        paddingTop: 12,
        borderTop: '1px solid var(--border-subtle)',
        display: 'flex',
        alignItems: 'center',
        gap: 8,
      }}>
        {/* Priority selector */}
        <button
          onClick={cyclePriority}
          title={`Priority: ${newPriority}`}
          style={{
            width: 20,
            height: 20,
            borderRadius: '50%',
            flexShrink: 0,
            background: PRIORITY_COLOR[newPriority],
            border: 'none',
            cursor: 'pointer',
            padding: 0,
          }}
        />

        {/* Text input */}
        <input
          type="text"
          value={newText}
          onChange={(e) => setNewText(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
          placeholder="Add a task..."
          style={{
            flex: 1,
            background: 'transparent',
            border: 'none',
            outline: 'none',
            fontSize: 13,
            color: 'var(--text-primary)',
            caretColor: 'var(--text-accent)',
          }}
        />

        {/* Submit button — only visible when input has text */}
        {newText.trim() && (
          <button
            onClick={handleAdd}
            style={{
              background: 'rgba(99,102,241,0.15)',
              border: '1px solid rgba(99,102,241,0.3)',
              borderRadius: 6,
              padding: '3px 10px',
              fontSize: 12,
              color: 'var(--text-accent)',
              cursor: 'pointer',
              flexShrink: 0,
            }}
          >
            Add
          </button>
        )}
      </div>
    </Card>
  );
}
