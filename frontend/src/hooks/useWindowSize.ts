import { useState, useEffect } from 'react';

export type Breakpoint = 'mobile' | 'tablet' | 'desktop';

export interface WindowSize {
  width: number;
  height: number;
  breakpoint: Breakpoint;
}

function getBreakpoint(width: number): Breakpoint {
  if (width < 768) return 'mobile';
  if (width < 1024) return 'tablet';
  return 'desktop';
}

export function useWindowSize(): WindowSize {
  const [size, setSize] = useState<WindowSize>(() => {
    const w = typeof window !== 'undefined' ? window.innerWidth : 1280;
    const h = typeof window !== 'undefined' ? window.innerHeight : 800;
    return { width: w, height: h, breakpoint: getBreakpoint(w) };
  });

  useEffect(() => {
    function handleResize() {
      const w = window.innerWidth;
      const h = window.innerHeight;
      setSize({ width: w, height: h, breakpoint: getBreakpoint(w) });
    }

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return size;
}
