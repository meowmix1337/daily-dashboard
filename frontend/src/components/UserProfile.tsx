import React, { useEffect, useRef, useState } from 'react';
import type { User } from '../types/auth';

interface Props {
  user: User;
}

function getInitials(name: string): string {
  if (!name.trim()) return '?';
  return name.trim().split(' ').filter(Boolean).map((n) => n[0]).join('').toUpperCase().slice(0, 2);
}

async function signOut(): Promise<void> {
  try {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' });
  } finally {
    window.location.href = '/login';
  }
}

export function UserProfile({ user }: Props): React.ReactElement {
  const [open, setOpen] = useState(false);
  const [imgError, setImgError] = useState(false);
  const [signOutHovered, setSignOutHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click or Escape
  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false);
    }
    document.addEventListener('mousedown', handleClick);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [open]);

  const showImg = !!user.avatar_url && !imgError;

  const avatarStyle = {
    width: 36,
    height: 36,
    borderRadius: '50%',
    border: '1px solid rgba(255,255,255,0.12)',
    cursor: 'pointer',
    background: 'rgba(99,102,241,0.25)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 13,
    fontWeight: 600,
    color: '#c7d2fe',
    overflow: 'hidden',
    flexShrink: 0,
    outline: 'none',
  } satisfies React.CSSProperties;

  const dropdownStyle = {
    position: 'absolute',
    top: 'calc(100% + 10px)',
    right: 0,
    width: 220,
    background: 'rgba(15,15,25,0.95)',
    border: '1px solid rgba(255,255,255,0.06)',
    borderRadius: 16,
    backdropFilter: 'blur(20px)',
    boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
    zIndex: 100,
    overflow: 'hidden',
  } satisfies React.CSSProperties;

  return (
    <div ref={containerRef} style={{ position: 'relative', display: 'inline-block', zIndex: 200 }}>
      <button
        type="button"
        aria-label="User profile"
        aria-expanded={open}
        aria-haspopup="true"
        onClick={() => setOpen((v) => !v)}
        style={avatarStyle}
      >
        {showImg ? (
          <img
            src={user.avatar_url}
            alt={user.name}
            width={36}
            height={36}
            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
            onError={() => setImgError(true)}
          />
        ) : (
          getInitials(user.name)
        )}
      </button>

      {open && (
        <div role="menu" style={dropdownStyle}>
          {/* User info */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '14px 16px',
            borderBottom: '1px solid rgba(255,255,255,0.06)',
          }}>
            <div style={{ ...avatarStyle, cursor: 'default', flexShrink: 0 }}>
              {showImg ? (
                <img
                  src={user.avatar_url}
                  alt={user.name}
                  width={36}
                  height={36}
                  style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                />
              ) : (
                getInitials(user.name)
              )}
            </div>
            <div style={{ minWidth: 0 }}>
              <div style={{
                fontSize: 13,
                fontWeight: 600,
                color: '#e2e2e8',
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
              }}>
                {user.name}
              </div>
              <div style={{
                fontSize: 11,
                color: '#6b7280',
                marginTop: 2,
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
              }}>
                {user.email}
              </div>
            </div>
          </div>

          {/* Sign out */}
          <button
            type="button"
            role="menuitem"
            onClick={() => { void signOut(); }}
            onMouseEnter={() => setSignOutHovered(true)}
            onMouseLeave={() => setSignOutHovered(false)}
            style={{
              width: '100%',
              textAlign: 'left',
              padding: '10px 16px',
              fontSize: 13,
              color: signOutHovered ? '#e2e2e8' : '#9ca3af',
              background: signOutHovered ? 'rgba(255,255,255,0.04)' : 'none',
              border: 'none',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              transition: 'color 0.15s, background 0.15s',
            }}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" y1="12" x2="9" y2="12" />
            </svg>
            Sign out
          </button>
        </div>
      )}
    </div>
  );
}
