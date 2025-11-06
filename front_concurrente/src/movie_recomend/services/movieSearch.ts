// Servicio para buscar películas en el CSV local y obtener imágenes de TMDB

import { convertTMDbMovie, TMDB_CONFIG, normalizeTMDBRating } from './tmdb';
import type { TMDbMovie } from './tmdb';

export interface MovieSearchResult {
  movieId: number;
  title: string;
  genres: string;
  poster?: string;
  rating?: number; // Rating de TMDB (0-10)
  csvRating?: number; // Rating normalizado de TMDB (0-5)
  overview?: string;
  releaseDate?: string;
}

let moviesCache: MovieSearchResult[] | null = null;

// Cargar todas las películas del CSV
export const loadMoviesFromCSV = async (): Promise<MovieSearchResult[]> => {
  if (moviesCache) {
    return moviesCache;
  }

  try {
    const response = await fetch('/data/movies.csv');
    const text = await response.text();
    const lines = text.split('\n').slice(1); // Saltar el header
    
    const movies: MovieSearchResult[] = lines
      .filter(line => line.trim())
      .map(line => {
        // Manejar comas dentro de las comillas
        const parts: string[] = [];
        let current = '';
        let inQuotes = false;

        for (let i = 0; i < line.length; i++) {
          const char = line[i];
          if (char === '"') {
            inQuotes = !inQuotes;
          } else if (char === ',' && !inQuotes) {
            parts.push(current);
            current = '';
          } else {
            current += char;
          }
        }
        parts.push(current);

        if (parts.length >= 3) {
          const movieId = parseInt(parts[0]) || 0;
          const title = parts[1].replace(/^"|"$/g, ''); // Remover comillas
          const genres = parts[2] || '';
          
          return {
            movieId,
            title,
            genres,
          };
        }
        return null;
      })
      .filter((movie): movie is MovieSearchResult => movie !== null && movie.movieId > 0);

    moviesCache = movies;
    return movies;
  } catch (error) {
    console.error('Error loading movies from CSV:', error);
    return [];
  }
};

// Buscar imágenes de películas en TMDB por título
const searchMovieInTMDB = async (title: string): Promise<MovieSearchResult | null> => {
  try {
    // Limpiar el título para mejor búsqueda (remover año si está presente)
    const cleanTitle = title.replace(/\s*\(\d{4}\)\s*$/, '').trim();
    
    const response = await fetch(
      `${TMDB_CONFIG.BASE_URL}/search/movie?api_key=${TMDB_CONFIG.API_KEY}&language=es-ES&query=${encodeURIComponent(cleanTitle)}&page=1`
    );
    
    if (!response.ok) {
      console.warn(`TMDB API error for "${title}":`, response.status);
      return null;
    }
    
    const data = await response.json();
    if (data.results && data.results.length > 0) {
      // Buscar la coincidencia más exacta (comparar títulos sin año)
      const cleanTitleLower = cleanTitle.toLowerCase();
      const exactMatch = data.results.find((movie: TMDbMovie) => {
        const movieTitle = movie.title.replace(/\s*\(\d{4}\)\s*$/, '').trim().toLowerCase();
        return movieTitle === cleanTitleLower;
      });
      
      // Si no hay coincidencia exacta, buscar la más similar
      const match = exactMatch || data.results.find((movie: TMDbMovie) => {
        const movieTitle = movie.title.replace(/\s*\(\d{4}\)\s*$/, '').trim().toLowerCase();
        return movieTitle.includes(cleanTitleLower) || cleanTitleLower.includes(movieTitle);
      }) || data.results[0];
      
      const movie = convertTMDbMovie(match);
      
      return {
        movieId: movie.movieId,
        title: movie.title,
        genres: movie.genres,
        poster: movie.poster,
        rating: movie.rating,
        overview: movie.overview,
        releaseDate: movie.releaseDate,
      };
    }
    return null;
  } catch (error) {
    console.error(`Error searching movie "${title}" in TMDB:`, error);
    return null;
  }
};

// Búsqueda inteligente de películas con imágenes de TMDB
export const searchMovies = async (
  query: string,
  limit: number = 20
): Promise<MovieSearchResult[]> => {
  if (!query || query.trim().length < 2) {
    return [];
  }

  if (!TMDB_CONFIG.API_KEY || TMDB_CONFIG.API_KEY === 'YOUR_API_KEY_HERE') {
    console.warn('TMDB API Key no configurada. Buscando solo en CSV local.');
    // Fallback a búsqueda solo en CSV si no hay API key
    return searchMoviesCSVOnly(query, limit);
  }

  const movies = await loadMoviesFromCSV();
  const searchQuery = query.toLowerCase().trim();

  // Buscar coincidencias exactas primero, luego parciales
  const exactMatches: MovieSearchResult[] = [];
  const partialMatches: MovieSearchResult[] = [];
  const fuzzyMatches: MovieSearchResult[] = [];

  for (const movie of movies) {
    const titleLower = movie.title.toLowerCase();
    const words = titleLower.split(/[\s,()\-]+/);

    // Coincidencia exacta
    if (titleLower === searchQuery) {
      exactMatches.push(movie);
    }
    // Coincidencia que empieza con la búsqueda
    else if (titleLower.startsWith(searchQuery)) {
      partialMatches.push(movie);
    }
    // Coincidencia que contiene la búsqueda
    else if (titleLower.includes(searchQuery)) {
      partialMatches.push(movie);
    }
    // Coincidencia difusa (palabras individuales)
    else if (words.some(word => word.startsWith(searchQuery) || searchQuery.includes(word))) {
      fuzzyMatches.push(movie);
    }
  }

  // Combinar resultados con prioridad
  const csvResults = [
    ...exactMatches,
    ...partialMatches,
    ...fuzzyMatches,
  ].slice(0, limit);

  // Buscar imágenes en TMDB para cada resultado
  const resultsWithImages = await Promise.all(
    csvResults.map(async (movie) => {
      // Buscar en TMDB por título
      const tmdbMovie = await searchMovieInTMDB(movie.title);
      
      if (tmdbMovie) {
        // Normalizar rating de TMDB (0-10) a escala 0-5
        const normalizedRating = tmdbMovie.rating !== undefined && tmdbMovie.rating > 0
          ? normalizeTMDBRating(tmdbMovie.rating)
          : undefined;
        
        // Combinar datos del CSV con imágenes de TMDB y rating de TMDB normalizado
        return {
          ...movie,
          poster: tmdbMovie.poster,
          rating: tmdbMovie.rating, // Rating de TMDB (0-10)
          csvRating: normalizedRating, // Rating normalizado de TMDB (0-5)
          overview: tmdbMovie.overview,
          releaseDate: tmdbMovie.releaseDate,
        };
      }
      
      // Si no hay datos de TMDB, mantener datos del CSV
      return movie;
    })
  );

  return resultsWithImages;
};

// Búsqueda solo en CSV (fallback sin API key)
const searchMoviesCSVOnly = async (
  query: string,
  limit: number = 20
): Promise<MovieSearchResult[]> => {
  const movies = await loadMoviesFromCSV();
  const searchQuery = query.toLowerCase().trim();

  const exactMatches: MovieSearchResult[] = [];
  const partialMatches: MovieSearchResult[] = [];
  const fuzzyMatches: MovieSearchResult[] = [];

  for (const movie of movies) {
    const titleLower = movie.title.toLowerCase();
    const words = titleLower.split(/[\s,()\-]+/);

    if (titleLower === searchQuery) {
      exactMatches.push(movie);
    } else if (titleLower.startsWith(searchQuery)) {
      partialMatches.push(movie);
    } else if (titleLower.includes(searchQuery)) {
      partialMatches.push(movie);
    } else if (words.some(word => word.startsWith(searchQuery) || searchQuery.includes(word))) {
      fuzzyMatches.push(movie);
    }
  }

  return [
    ...exactMatches,
    ...partialMatches,
    ...fuzzyMatches,
  ].slice(0, limit);
};

