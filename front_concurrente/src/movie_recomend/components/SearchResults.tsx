import React from 'react';
import { MovieSearchResult } from '../services/movieSearch';
import MovieCard from './MovieCard';
import { Movie } from '../services/tmdb';

interface SearchResultsProps {
  results: MovieSearchResult[];
  onClose: () => void;
  onMovieSelect?: (movie: MovieSearchResult) => void;
}

const SearchResults: React.FC<SearchResultsProps> = ({ 
  results, 
  onClose,
  onMovieSelect 
}) => {
  if (results.length === 0) {
    return null;
  }

  // Convertir MovieSearchResult a Movie para usar con MovieCard
  const convertToMovie = (result: MovieSearchResult): Movie => {
    return {
      movieId: result.movieId,
      title: result.title,
      genres: result.genres,
      poster: result.poster,
      rating: result.rating,
      csvRating: result.csvRating, // Rating normalizado de TMDB (0-5)
      overview: result.overview,
      releaseDate: result.releaseDate,
    };
  };

  const styles = {
    overlay: {
      position: 'fixed' as const,
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'rgba(0, 0, 0, 0.8)',
      zIndex: 1000,
      display: 'flex',
      alignItems: 'flex-start',
      justifyContent: 'center',
      paddingTop: '100px',
      padding: '100px 20px 20px',
      overflowY: 'auto' as const,
    },
    resultsContainer: {
      background: '#1a1825',
      borderRadius: '16px',
      padding: '24px',
      maxWidth: '1200px',
      width: '100%',
      maxHeight: '80vh',
      overflowY: 'auto' as const,
      border: '1px solid #2a2640',
      boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)',
      position: 'relative' as const,
    },
    header: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginBottom: '24px',
    },
    title: {
      fontSize: '24px',
      fontWeight: '600' as const,
      color: '#ffffff',
      margin: 0,
    },
    closeButton: {
      background: '#262335',
      border: '1px solid #2a2640',
      borderRadius: '8px',
      color: '#ffffff',
      width: '32px',
      height: '32px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
    },
    resultsCount: {
      color: '#8b87a0',
      fontSize: '14px',
      marginBottom: '20px',
    },
    grid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
      gap: '20px',
    },
    emptyState: {
      textAlign: 'center' as const,
      color: '#8b87a0',
      fontSize: '16px',
      padding: '40px',
    },
  };

  return (
    <div 
      style={styles.overlay}
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <div style={styles.resultsContainer}>
        <div style={styles.header}>
          <h2 style={styles.title}>Resultados de búsqueda</h2>
          <button
            style={styles.closeButton}
            onClick={onClose}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = '#2a2640';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#262335';
            }}
          >
            ✕
          </button>
        </div>
        <div style={styles.resultsCount}>
          {results.length} {results.length === 1 ? 'película encontrada' : 'películas encontradas'}
        </div>
        <div style={styles.grid}>
          {results.map((result) => (
            <div
              key={result.movieId}
              onClick={() => {
                if (onMovieSelect) {
                  onMovieSelect(result);
                }
                onClose();
              }}
            >
              <MovieCard 
                movie={convertToMovie(result)} 
                size="small"
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default SearchResults;

