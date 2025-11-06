import React, { useState, useEffect } from 'react';
import { Film, User, LogOut } from 'lucide-react';
import { 
  Movie, 
  fetchTrendingMovies, 
  fetchTopRatedMovies, 
  fetchPopularMovies,
  isApiKeyConfigured 
} from '../services/tmdb';
import MovieCard from '../components/MovieCard';
import ScrollableSection from '../components/ScrollableSection';
import SearchBar from '../components/SearchBar';
import SearchResults from '../components/SearchResults';
import { searchMovies, MovieSearchResult } from '../services/movieSearch';

interface DashboardProps {
  onLogout?: () => void;
}

const Dashboard: React.FC<DashboardProps> = ({ onLogout }) => {
  const [recommendedMovies, setRecommendedMovies] = useState<Movie[]>([]);
  const [topIMDBMovies, setTopIMDBMovies] = useState<Movie[]>([]);
  const [topTMDbMovies, setTopTMDbMovies] = useState<Movie[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [userName, setUserName] = useState<string>('Usuario');
  const [searchResults, setSearchResults] = useState<MovieSearchResult[]>([]);
  const [showSearchResults, setShowSearchResults] = useState(false);
  const [isManuallyClosed, setIsManuallyClosed] = useState(false);
  const [lastSearchQuery, setLastSearchQuery] = useState<string>('');

  // Obtener el nombre del usuario del localStorage
  useEffect(() => {
    const userEmail = localStorage.getItem('userEmail');
    if (userEmail) {
      // Extraer el nombre del email (antes del @) o usar el email completo
      const name = userEmail.split('@')[0];
      setUserName(name.charAt(0).toUpperCase() + name.slice(1));
    }
  }, []);

  const handleSearch = async (query: string) => {
    // Si el query cambió, resetear el estado de cierre manual
    if (query !== lastSearchQuery) {
      setIsManuallyClosed(false);
      setLastSearchQuery(query);
    }

    if (!query || query.trim().length < 2) {
      setSearchResults([]);
      setShowSearchResults(false);
      return;
    }

    // Si el modal fue cerrado manualmente, no abrirlo automáticamente
    if (isManuallyClosed) {
      return;
    }

    try {
      const results = await searchMovies(query, 30);
      setSearchResults(results);
      // Solo mostrar resultados si hay resultados y no fue cerrado manualmente
      if (results.length > 0 && !isManuallyClosed) {
        setShowSearchResults(true);
      }
    } catch (error) {
      console.error('Error searching movies:', error);
      setSearchResults([]);
      setShowSearchResults(false);
    }
  };

  const handleCloseSearch = () => {
    setShowSearchResults(false);
    setIsManuallyClosed(true);
  };

  const handleMovieSelect = (movie: MovieSearchResult) => {
    console.log('Película seleccionada:', movie);
    // Aquí puedes navegar a una página de detalles o hacer algo con la película seleccionada
  };

  const handleLogout = () => {
    if (onLogout) {
      onLogout();
    }
  };

  // Cargar todos los datos de TMDB
  useEffect(() => {
    const loadAllMovies = async () => {
      setIsLoading(true);
      try {
        if (isApiKeyConfigured()) {
          const [trending, topRated, popular] = await Promise.all([
            fetchTrendingMovies(),
            fetchTopRatedMovies(),
            fetchPopularMovies(),
          ]);

          // Los ratings ya vienen normalizados de TMDB en csvRating
          setRecommendedMovies(trending);
          setTopIMDBMovies(topRated);
          setTopTMDbMovies(popular);
        } else {
          console.warn('TMDB API Key no configurada. Por favor, configura VITE_TMDB_API_KEY en tu archivo .env');
        }
      } catch (error) {
        console.error('Error loading movies:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadAllMovies();
  }, []);

  const styles = {
    container: {
      minHeight: '100vh',
      width: '100%',
      background: '#0f0f0f',
      color: '#ffffff',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
      padding: '0',
      margin: '0',
    },
    header: {
      padding: '20px 40px',
      background: '#1a1825',
      borderBottom: '1px solid #2a2640',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: '20px',
    },
    logo: {
      display: 'flex',
      alignItems: 'center',
      gap: '10px',
      color: 'white',
      flexShrink: 0,
    },
    logoText: {
      fontSize: '24px',
      fontWeight: 'bold',
      letterSpacing: '-0.5px',
    },
    searchContainer: {
      flex: 1,
      display: 'flex',
      justifyContent: 'center',
      maxWidth: '600px',
      margin: '0 auto',
    },
    userSection: {
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      flexShrink: 0,
    },
    userInfo: {
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      color: '#ffffff',
      fontSize: '14px',
    },
    userName: {
      color: '#ffffff',
      fontSize: '14px',
      fontWeight: '500' as const,
    },
    logoutButton: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '8px',
      background: '#262335',
      border: '1px solid #2a2640',
      borderRadius: '8px',
      color: '#ffffff',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      width: '40px',
      height: '40px',
    },
    content: {
      padding: '40px',
      maxWidth: '1400px',
      margin: '0 auto',
    },
  };

  const loadingStyle = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: '80vh',
    color: '#8b87a0',
    fontSize: '18px',
  };

  if (isLoading) {
    return (
      <div style={styles.container}>
        <div style={styles.header}>
          <div style={styles.logo}>
            <Film size={32} />
            <span style={styles.logoText}>MOVIEFLIX</span>
          </div>
          <div style={styles.searchContainer}>
            <SearchBar 
              placeholder="Buscar películas..."
              onSearch={handleSearch}
              onQueryChange={(query) => {
                if (query !== lastSearchQuery) {
                  setIsManuallyClosed(false);
                  setLastSearchQuery(query);
                }
              }}
            />
          </div>
          <div style={styles.userSection}>
            <div style={styles.userInfo}>
              <User size={20} />
              <span style={styles.userName}>{userName}</span>
            </div>
            <button
              style={styles.logoutButton}
              onClick={handleLogout}
              title="Logout"
            >
              <LogOut size={20} />
            </button>
          </div>
        </div>
        <div style={loadingStyle}>
          Cargando películas...
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <div style={styles.logo}>
          <Film size={32} />
          <span style={styles.logoText}>MOVIEFLIX</span>
        </div>
        
        <div style={styles.searchContainer}>
          <SearchBar 
            placeholder="Buscar películas..."
            onSearch={handleSearch}
            onQueryChange={(query) => {
              // Resetear el estado de cierre manual cuando el usuario escribe algo nuevo
              if (query !== lastSearchQuery) {
                setIsManuallyClosed(false);
                setLastSearchQuery(query);
              }
            }}
          />
        </div>

        <div style={styles.userSection}>
          <div style={styles.userInfo}>
            <User size={20} />
            <span style={styles.userName}>{userName}</span>
          </div>
          <button
            style={styles.logoutButton}
            onClick={handleLogout}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = '#2a2640';
              e.currentTarget.style.borderColor = '#667eea';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#262335';
              e.currentTarget.style.borderColor = '#2a2640';
            }}
            title="Logout"
          >
            <LogOut size={20} />
          </button>
        </div>
      </div>

      {/* Content */}
      <div style={styles.content}>
        {/* 1. Recommended Section */}
        <ScrollableSection 
          title="Recommended" 
          sectionId="recommended-grid"
          size="large"
        >
          {recommendedMovies.map((movie) => (
            <MovieCard 
              key={movie.movieId} 
              movie={movie} 
              size="large"
            />
          ))}
        </ScrollableSection>

        {/* 2. Top in IMDB Section */}
        <ScrollableSection 
          title="Top in IMDB" 
          sectionId="imdb-grid"
          size="small"
        >
          {topIMDBMovies.map((movie) => (
            <MovieCard 
              key={movie.movieId} 
              movie={movie} 
              size="small"
            />
          ))}
        </ScrollableSection>

        {/* 3. Top in TMDB Section */}
        <ScrollableSection 
          title="Top in TMDB" 
          sectionId="tmdb-grid"
          size="small"
        >
          {topTMDbMovies.map((movie) => (
            <MovieCard 
              key={movie.movieId} 
              movie={movie} 
              size="small"
            />
          ))}
        </ScrollableSection>
      </div>

      {/* Search Results Modal */}
      {showSearchResults && (
        <SearchResults
          results={searchResults}
          onClose={handleCloseSearch}
          onMovieSelect={handleMovieSelect}
        />
      )}
    </div>
  );
};

export default Dashboard;

