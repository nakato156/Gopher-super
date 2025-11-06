import React from 'react';
import { Film } from 'lucide-react';
import { Movie } from '../services/tmdb';
import StarRating from './StarRating';

interface MovieCardProps {
  movie: Movie;
  size?: 'large' | 'small';
  onClick?: () => void;
}

const MovieCard: React.FC<MovieCardProps> = ({ movie, size = 'large', onClick }) => {
  const isLarge = size === 'large';

  const styles = {
    movieCard: {
      background: '#1a1825',
      borderRadius: '12px',
      overflow: 'hidden' as const,
      cursor: onClick ? 'pointer' as const : 'default' as const,
      transition: 'all 0.3s ease',
      border: '1px solid #2a2640',
      minWidth: isLarge ? '280px' : '200px',
    },
    moviePoster: {
      width: '100%',
      height: isLarge ? '420px' : '300px',
      objectFit: 'cover' as const,
      background: '#262335',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: '#8b87a0',
      fontSize: '14px',
    },
    movieInfo: {
      padding: '16px',
    },
    movieTitle: {
      fontSize: '16px',
      fontWeight: '500' as const,
      color: '#ffffff',
      marginBottom: '8px',
      overflow: 'hidden' as const,
      textOverflow: 'ellipsis' as const,
      whiteSpace: 'nowrap' as const,
    },
    movieGenres: {
      fontSize: '12px',
      color: '#8b87a0',
      overflow: 'hidden' as const,
      textOverflow: 'ellipsis' as const,
      whiteSpace: 'nowrap' as const,
      marginBottom: '8px',
    },
    ratingContainer: {
      display: 'flex',
      alignItems: 'center',
      gap: '6px',
    },
    ratingText: {
      fontSize: '12px',
      color: '#8b87a0',
    },
  };

  return (
    <div 
      style={styles.movieCard}
      onClick={onClick}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = 'translateY(-4px)';
        e.currentTarget.style.boxShadow = '0 8px 16px rgba(0, 0, 0, 0.3)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = 'translateY(0)';
        e.currentTarget.style.boxShadow = 'none';
      }}
    >
      <div style={styles.moviePoster}>
        {movie.poster ? (
          <img 
            src={movie.poster} 
            alt={movie.title} 
            style={styles.moviePoster}
            onError={(e) => {
              // Si falla la imagen, mostrar placeholder
              const target = e.target as HTMLImageElement;
              target.style.display = 'none';
              const parent = target.parentElement;
              if (parent) {
                parent.innerHTML = '<div style="display: flex; align-items: center; justify-content: center; width: 100%; height: 100%;"><svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"></rect><line x1="7" y1="2" x2="7" y2="22"></line><line x1="17" y1="2" x2="17" y2="22"></line><line x1="2" y1="12" x2="22" y2="12"></line><line x1="2" y1="7" x2="7" y2="7"></line><line x1="2" y1="17" x2="7" y2="17"></line><line x1="17" y1="17" x2="22" y2="17"></line><line x1="17" y1="7" x2="22" y2="7"></line></svg></div>';
              }
            }}
          />
        ) : (
          <Film size={isLarge ? 48 : 36} />
        )}
      </div>
      <div style={styles.movieInfo}>
        <div style={styles.movieTitle}>{movie.title}</div>
        <div style={styles.movieGenres}>{movie.genres}</div>
        {movie.csvRating !== undefined && (
          <div style={styles.ratingContainer}>
            <StarRating rating={movie.csvRating} size={isLarge ? 16 : 14} />
            <span style={styles.ratingText}>
              {movie.csvRating.toFixed(1)}
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

export default MovieCard;

