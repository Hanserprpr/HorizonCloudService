import { useState, useEffect } from 'react';
import { useUIStore } from '@stores/uiStore';

interface BreakpointState {
  xs: boolean;  // < 576px
  sm: boolean;  // >= 576px
  md: boolean;  // >= 768px
  lg: boolean;  // >= 992px
  xl: boolean;  // >= 1200px
  xxl: boolean; // >= 1600px
}

const breakpoints = {
  xs: 0,
  sm: 576,
  md: 768,
  lg: 992,
  xl: 1200,
  xxl: 1600,
};

export const useResponsive = () => {
  const { setIsMobile } = useUIStore();
  
  const [screenSize, setScreenSize] = useState<BreakpointState>(() => {
    if (typeof window === 'undefined') {
      return {
        xs: false,
        sm: true,
        md: true,
        lg: true,
        xl: true,
        xxl: false,
      };
    }
    
    const width = window.innerWidth;
    return {
      xs: width < breakpoints.sm,
      sm: width >= breakpoints.sm,
      md: width >= breakpoints.md,
      lg: width >= breakpoints.lg,
      xl: width >= breakpoints.xl,
      xxl: width >= breakpoints.xxl,
    };
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleResize = () => {
      const width = window.innerWidth;
      const newState = {
        xs: width < breakpoints.sm,
        sm: width >= breakpoints.sm,
        md: width >= breakpoints.md,
        lg: width >= breakpoints.lg,
        xl: width >= breakpoints.xl,
        xxl: width >= breakpoints.xxl,
      };
      
      setScreenSize(newState);
      
      // 更新移动端状态
      const isMobile = width < breakpoints.md;
      setIsMobile(isMobile);
    };

    // 初始化时设置移动端状态
    const initialIsMobile = window.innerWidth < breakpoints.md;
    setIsMobile(initialIsMobile);

    window.addEventListener('resize', handleResize);
    
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [setIsMobile]);

  const isMobile = screenSize.xs || !screenSize.md;

  return {
    ...screenSize,
    isMobile,
    isDesktop: screenSize.lg,
    isTablet: screenSize.md && !screenSize.lg,
  };
};