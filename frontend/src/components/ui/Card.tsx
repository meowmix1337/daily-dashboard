import React, { useEffect, useState } from 'react';

interface CardProps {
  children: React.ReactNode;
  delay?: number;
  span?: number;
  className?: string;
  noGridSpan?: boolean;
  style?: React.CSSProperties;
}

export function Card({ children, delay = 0, span = 1, className = '', noGridSpan = false, style }: CardProps): React.ReactElement {
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setLoaded(true), 100 + delay * 1000);
    return () => clearTimeout(timer);
  }, [delay]);

  return (
    <div
      style={{
        ...(noGridSpan ? {} : { gridColumn: `span ${span}` }),
        background: 'var(--bg-card)',
        border: '1px solid var(--bg-card-border)',
        borderRadius: 16,
        padding: 24,
        backdropFilter: 'blur(20px)',
        opacity: loaded ? 1 : 0,
        transform: loaded ? 'translateY(0)' : 'translateY(16px)',
        transition: `opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1) ${delay}s, transform 0.7s cubic-bezier(0.16, 1, 0.3, 1) ${delay}s`,
        ...style,
      }}
      className={className}
    >
      {children}
    </div>
  );
}
