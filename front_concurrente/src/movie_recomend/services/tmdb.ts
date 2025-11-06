// Servicio para interactuar con la API de TMDB

export interface TMDbMovie {
  id: number;
  title: string;
  overview: string;
  poster_path: string | null;
  release_date: string;
  vote_average: number;
  genre_ids: number[];
  backdrop_path: string | null;
}

export interface Movie {
  movieId: number;
  title: string;
  genres: string;
  poster?: string;
  rating?: number; // Rating de TMDB (0-10)
  csvRating?: number; // Rating promedio del CSV (1-5)
  overview?: string;
  releaseDate?: string;
}

// Mapear géneros de TMDB
export const genreMap: { [key: number]: string } = {
  28: 'Action',
  12: 'Adventure',
  16: 'Animation',
  35: 'Comedy',
  80: 'Crime',
  99: 'Documentary',
  18: 'Drama',
  10751: 'Family',
  14: 'Fantasy',
  36: 'History',
  27: 'Horror',
  10402: 'Music',
  9648: 'Mystery',
  10749: 'Romance',
  878: 'Science Fiction',
  10770: 'TV Movie',
  53: 'Thriller',
  10752: 'War',
  37: 'Western',
};

// Configuración de TMDB
export const TMDB_CONFIG = {
  API_KEY: import.meta.env.VITE_TMDB_API_KEY || 'YOUR_API_KEY_HERE',
  BASE_URL: 'https://api.themoviedb.org/3',
  IMAGE_BASE_URL: 'https://image.tmdb.org/t/p/w500',
};

// Función para obtener la URL de la imagen del poster
export const getPosterUrl = (posterPath: string | null | undefined): string | undefined => {
  if (!posterPath) return undefined;
  return `${TMDB_CONFIG.IMAGE_BASE_URL}${posterPath}`;
};

// Función para convertir géneros de IDs a string
export const convertGenres = (genreIds: number[]): string => {
  const genres = genreIds
    .map(id => genreMap[id] || '')
    .filter(Boolean)
    .join('|');
  return genres || 'Unknown';
};

// Normalizar rating de TMDB (0-10) a escala 0-5
export const normalizeTMDBRating = (tmdbRating: number): number => {
  // TMDB rating está en escala 0-10, normalizar a 0-5
  return (tmdbRating / 10) * 5;
};

// Función para convertir película de TMDB al formato local
export const convertTMDbMovie = (tmdbMovie: TMDbMovie): Movie => {
  // Normalizar el rating de TMDB (0-10) a escala 0-5
  const normalizedRating = normalizeTMDBRating(tmdbMovie.vote_average);
  
  return {
    movieId: tmdbMovie.id,
    title: tmdbMovie.title,
    genres: convertGenres(tmdbMovie.genre_ids),
    poster: getPosterUrl(tmdbMovie.poster_path),
    rating: tmdbMovie.vote_average, // Rating original de TMDB (0-100)
    csvRating: normalizedRating, // Rating normalizado de TMDB (0-5)
    overview: tmdbMovie.overview,
    releaseDate: tmdbMovie.release_date,
  };
};

// Función para obtener películas en tendencia
export const fetchTrendingMovies = async (): Promise<Movie[]> => {
  try {
    const response = await fetch(
      `${TMDB_CONFIG.BASE_URL}/trending/movie/day?api_key=${TMDB_CONFIG.API_KEY}&language=es-ES`
    );
    if (!response.ok) throw new Error('Failed to fetch trending movies');
    const data = await response.json();
    return data.results.map(convertTMDbMovie);
  } catch (error) {
    console.error('Error loading trending movies:', error);
    throw error;
  }
};

// Función para obtener películas mejor valoradas
export const fetchTopRatedMovies = async (page: number = 1): Promise<Movie[]> => {
  try {
    const response = await fetch(
      `${TMDB_CONFIG.BASE_URL}/movie/top_rated?api_key=${TMDB_CONFIG.API_KEY}&language=es-ES&page=${page}`
    );
    if (!response.ok) throw new Error('Failed to fetch top rated movies');
    const data = await response.json();
    return data.results.map(convertTMDbMovie);
  } catch (error) {
    console.error('Error loading top rated movies:', error);
    throw error;
  }
};

// Función para obtener películas populares
export const fetchPopularMovies = async (page: number = 1): Promise<Movie[]> => {
  try {
    const response = await fetch(
      `${TMDB_CONFIG.BASE_URL}/movie/popular?api_key=${TMDB_CONFIG.API_KEY}&language=es-ES&page=${page}`
    );
    if (!response.ok) throw new Error('Failed to fetch popular movies');
    const data = await response.json();
    return data.results.map(convertTMDbMovie);
  } catch (error) {
    console.error('Error loading popular movies:', error);
    throw error;
  }
};

// Función para verificar si la API key está configurada
export const isApiKeyConfigured = (): boolean => {
  return TMDB_CONFIG.API_KEY !== 'YOUR_API_KEY_HERE' && !!TMDB_CONFIG.API_KEY;
};

