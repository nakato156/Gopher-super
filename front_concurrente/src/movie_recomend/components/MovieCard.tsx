import React, { useState, useEffect } from 'react';
import { Film, ListPlus, Check } from 'lucide-react';
import { Movie } from '../services/tmdb';
import StarRating from './StarRating';
import { addMovieToList, removeMovieFromList, isMovieInList } from '../services/userList';

interface MovieCardProps {
  movie: Movie;
  size?: 'large' | 'small';
  onClick?: () => void;
}

const MovieCard: React.FC<MovieCardProps> = ({ movie, size = 'large', onClick }) => {
  const isLarge = size === 'large';
  const [inList, setInList] = useState(false);

  useEffect(() => {
    setInList(isMovieInList(movie.movieId));
  }, [movie.movieId]);

  const handleAddToList = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (inList) {
      removeMovieFromList(movie.movieId);
      setInList(false);
    } else {
      addMovieToList(movie);
      setInList(true);
    }
  };

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
    posterContainer: {
      position: 'relative' as const,
      width: '100%',
      height: isLarge ? '420px' : '300px',
    },
    moviePoster: {
      width: '100%',
      height: '100%',
      objectFit: 'cover' as const,
      background: '#262335',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: '#8b87a0',
      fontSize: '14px',
    },
    addToListButton: {
      position: 'absolute' as const,
      top: '12px',
      right: '12px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '8px',
      background: 'rgba(26, 24, 37, 0.9)',
      border: '1px solid #2a2640',
      borderRadius: '8px',
      color: '#ffffff',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      width: '36px',
      height: '36px',
      zIndex: 10,
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
      <div style={styles.posterContainer}>
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
        <button
          style={{
            ...styles.addToListButton,
            background: inList ? 'rgba(102, 126, 234, 0.9)' : 'rgba(26, 24, 37, 0.9)',
            borderColor: inList ? '#667eea' : '#2a2640',
          }}
          onClick={handleAddToList}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = inList 
              ? 'rgba(118, 75, 162, 0.95)' 
              : 'rgba(42, 38, 64, 0.95)';
            e.currentTarget.style.borderColor = '#667eea';
            e.currentTarget.style.transform = 'scale(1.1)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = inList 
              ? 'rgba(102, 126, 234, 0.9)' 
              : 'rgba(26, 24, 37, 0.9)';
            e.currentTarget.style.borderColor = inList ? '#667eea' : '#2a2640';
            e.currentTarget.style.transform = 'scale(1)';
          }}
          title={inList ? 'Quitar de la lista' : 'Agregar a la lista'}
        >
          {inList ? (
            <Check size={isLarge ? 20 : 18} />
          ) : (
            <ListPlus size={isLarge ? 20 : 18} />
          )}
        </button>
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

