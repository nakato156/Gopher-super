import React from 'react';
import { Star } from 'lucide-react';

interface StarRatingProps {
  rating: number; // Rating de 0 a 5
  size?: number;
}

const StarRating: React.FC<StarRatingProps> = ({ rating, size = 14 }) => {
  // Asegurar que el rating esté entre 0 y 5
  const normalizedRating = Math.max(0, Math.min(5, rating));
  const fullStars = Math.floor(normalizedRating);
  const remainder = normalizedRating % 1;
  
  // Si hay decimal (> 0 y < 1), mostrar estrella parcial
  const hasPartialStar = remainder > 0 && remainder < 1;
  const partialStarPercentage = remainder * 100; // Porcentaje de llenado (0-100%)
  const emptyStars = 5 - fullStars - (hasPartialStar ? 1 : 0);

  const styles = {
    container: {
      display: 'flex',
      alignItems: 'center',
      gap: '2px',
    },
    star: {
      color: '#667eea',
      fill: '#667eea',
    },
    starEmpty: {
      color: '#2a2640',
      fill: 'none',
      stroke: '#2a2640',
    },
    partialStarContainer: {
      position: 'relative' as const,
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: `${size}px`,
      height: `${size}px`,
      verticalAlign: 'middle',
    },
    partialStarFilled: {
      position: 'absolute' as const,
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      clipPath: `inset(0 ${100 - partialStarPercentage}% 0 0)`,
    },
    partialStarEmpty: {
      position: 'absolute' as const,
      top: 0,
      left: 0,
      width: '100%',
      height: '100%',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: -1,
    },
  };

  return (
    <div style={styles.container}>
      {/* Estrellas completas */}
      {Array.from({ length: fullStars }).map((_, index) => (
        <Star
          key={`full-${index}`}
          size={size}
          style={styles.star}
          fill="currentColor"
        />
      ))}
      
      {/* Estrella parcial (con decimal) */}
      {hasPartialStar && (
        <div style={styles.partialStarContainer}>
          {/* Parte llena según el decimal */}
          <div style={styles.partialStarFilled}>
            <Star
              size={size}
              style={styles.star}
              fill="currentColor"
            />
          </div>
          {/* Parte vacía */}
          <div style={styles.partialStarEmpty}>
            <Star
              size={size}
              style={styles.starEmpty}
            />
          </div>
        </div>
      )}
      
      {/* Estrellas vacías */}
      {Array.from({ length: emptyStars }).map((_, index) => (
        <Star
          key={`empty-${index}`}
          size={size}
          style={styles.starEmpty}
        />
      ))}
    </div>
  );
};

export default StarRating;

