import React, { ReactNode } from 'react';
import arrowForward from '../../assets/arrow_foward.svg';

interface ScrollableSectionProps {
  title: string;
  sectionId: string;
  children: ReactNode;
  size?: 'large' | 'small';
}

const ScrollableSection: React.FC<ScrollableSectionProps> = ({ 
  title, 
  sectionId, 
  children,
  size = 'large'
}) => {
  const scrollSection = (direction: 'left' | 'right') => {
    const section = document.getElementById(sectionId);
    if (section) {
      const scrollAmount = direction === 'left' ? -400 : 400;
      section.scrollBy({ left: scrollAmount, behavior: 'smooth' });
    }
  };

  const styles = {
    section: {
      marginBottom: '60px',
    },
    sectionHeader: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginBottom: '24px',
    },
    sectionTitle: {
      fontSize: '32px',
      fontWeight: '600' as const,
      color: '#ffffff',
      margin: '0',
    },
    scrollButtons: {
      display: 'flex',
      gap: '10px',
    },
    scrollButton: {
      background: '#262335',
      border: '1px solid #2a2640',
      borderRadius: '8px',
      color: '#ffffff',
      width: '40px',
      height: '40px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      padding: '0',
    },
    arrowIcon: {
      width: '20px',
      height: '20px',
      objectFit: 'contain' as const,
      filter: 'brightness(0) invert(1)',
    },
    arrowIconLeft: {
      width: '20px',
      height: '20px',
      objectFit: 'contain' as const,
      transform: 'rotate(180deg)',
      filter: 'brightness(0) invert(1)',
    },
    gridContainer: {
      position: 'relative' as const,
      overflow: 'hidden',
    },
    grid: {
      display: 'flex',
      gap: size === 'large' ? '24px' : '16px',
      overflowX: 'auto' as const,
      overflowY: 'hidden' as const,
      scrollBehavior: 'smooth' as const,
      paddingBottom: '10px',
      scrollbarWidth: 'thin' as const,
      scrollbarColor: '#2a2640 #1a1825',
    },
  };

  return (
    <div style={styles.section}>
      <div style={styles.sectionHeader}>
        <h2 style={styles.sectionTitle}>{title}</h2>
        <div style={styles.scrollButtons}>
          <button
            style={styles.scrollButton}
            onClick={() => scrollSection('left')}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = '#2a2640';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#262335';
            }}
          >
            <img 
              src={arrowForward} 
              alt="Scroll left" 
              style={styles.arrowIconLeft}
            />
          </button>
          <button
            style={styles.scrollButton}
            onClick={() => scrollSection('right')}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = '#2a2640';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#262335';
            }}
          >
            <img 
              src={arrowForward} 
              alt="Scroll right" 
              style={styles.arrowIcon}
            />
          </button>
        </div>
      </div>
      <div style={styles.gridContainer}>
        <div id={sectionId} style={styles.grid}>
          {children}
        </div>
      </div>
    </div>
  );
};

export default ScrollableSection;

