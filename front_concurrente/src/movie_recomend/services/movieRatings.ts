// Servicio para cargar y gestionar ratings de películas desde ratings.csv

export interface MovieRating {
  userId: number;
  movieId: number;
  rating: number;
  timestamp: number;
}

let ratingsCache: Map<number, number[]> | null = null; // movieId (CSV) -> array de ratings
let tmdbToMovieIdCache: Map<number, number> | null = null; // tmdbId -> movieId (CSV)

// Cargar mapeo de tmdbId a movieId del CSV desde links.csv
const loadTMDBToMovieIdMapping = async (): Promise<Map<number, number>> => {
  if (tmdbToMovieIdCache) {
    return tmdbToMovieIdCache;
  }

  try {
    const response = await fetch('/data/links.csv');
    const text = await response.text();
    const lines = text.split('\n').slice(1); // Saltar el header
    
    const mapping = new Map<number, number>();
    
    for (const line of lines) {
      if (!line.trim()) continue;
      
      const parts = line.split(',').map(p => p.trim());
      if (parts.length >= 3) {
        const movieId = parseInt(parts[0]) || 0;
        const tmdbId = parseInt(parts[2]) || 0;
        
        if (movieId > 0 && tmdbId > 0) {
          mapping.set(tmdbId, movieId);
        }
      }
    }
    
    console.log(`Cargado mapeo TMDB->MovieId: ${mapping.size} entradas`);

    tmdbToMovieIdCache = mapping;
    return mapping;
  } catch (error) {
    console.error('Error loading TMDB to MovieId mapping:', error);
    return new Map<number, number>();
  }
};

// Cargar ratings del CSV y calcular promedios por película
const loadRatingsFromCSV = async (): Promise<Map<number, number>> => {
  if (ratingsCache) {
    // Convertir el cache a promedios
    const averages = new Map<number, number>();
    ratingsCache.forEach((ratings, movieId) => {
      const average = ratings.reduce((sum, rating) => sum + rating, 0) / ratings.length;
      averages.set(movieId, average);
    });
    return averages;
  }

  try {
    console.log('Iniciando carga de ratings.csv...');
    const response = await fetch('/data/ratings.csv');
    console.log('Response status:', response.status, response.statusText);
    
    if (!response.ok) {
      console.error('Error al cargar ratings.csv:', response.status, response.statusText);
      return new Map<number, number>();
    }
    
    if (!response.body) {
      console.error('Response body is null');
      return new Map<number, number>();
    }
    
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    const ratingsMap = new Map<number, number[]>();
    let buffer = '';
    let processedLines = 0;
    let validRatings = 0;
    let isHeader = true;
    
    console.log('Procesando ratings.csv en streaming...');
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || ''; // Mantener la última línea incompleta en el buffer
      
      for (const line of lines) {
        if (!line.trim()) continue;
        
        if (isHeader) {
          isHeader = false;
          continue;
        }
        
        processedLines++;
        const parts = line.split(',').map(p => p.trim());
        
        if (parts.length >= 3) {
          const movieId = parseInt(parts[1]) || 0;
          const rating = parseFloat(parts[2]) || 0;
          
          if (movieId > 0 && rating >= 1 && rating <= 5) {
            validRatings++;
            if (!ratingsMap.has(movieId)) {
              ratingsMap.set(movieId, []);
            }
            ratingsMap.get(movieId)!.push(rating);
          }
        }
        
        if (processedLines % 100000 === 0) {
          console.log(`Procesadas ${processedLines} líneas, ${ratingsMap.size} películas únicas...`);
        }
      }
    }
    
    // Procesar la última línea del buffer si existe
    if (buffer.trim() && !isHeader) {
      processedLines++;
      const parts = buffer.split(',').map(p => p.trim());
      if (parts.length >= 3) {
        const movieId = parseInt(parts[1]) || 0;
        const rating = parseFloat(parts[2]) || 0;
        if (movieId > 0 && rating >= 1 && rating <= 5) {
          validRatings++;
          if (!ratingsMap.has(movieId)) {
            ratingsMap.set(movieId, []);
          }
          ratingsMap.get(movieId)!.push(rating);
        }
      }
    }
    
    console.log(`Cargado ratings: ${ratingsMap.size} películas, ${validRatings} ratings válidos de ${processedLines} líneas procesadas`);

    // Calcular promedios
    const averages = new Map<number, number>();
    ratingsMap.forEach((ratings, movieId) => {
      const average = ratings.reduce((sum, rating) => sum + rating, 0) / ratings.length;
      averages.set(movieId, average);
    });

    // Guardar en cache
    ratingsCache = ratingsMap;
    
    return averages;
  } catch (error) {
    console.error('Error loading ratings from CSV:', error);
    return new Map<number, number>();
  }
};

// Obtener el rating promedio de una película usando tmdbId
export const getMovieRatingByTMDBId = async (tmdbId: number): Promise<number | null> => {
  const mapping = await loadTMDBToMovieIdMapping();
  const csvMovieId = mapping.get(tmdbId);
  
  if (!csvMovieId) {
    return null;
  }
  
  const averages = await loadRatingsFromCSV();
  return averages.get(csvMovieId) || null;
};

// Normalizar rating de TMDB (0-10) a escala 0-5
export const normalizeTMDBRating = (tmdbRating: number): number => {
  // TMDB rating está en escala 0-10, normalizar a 0-5
  return (tmdbRating / 10) * 5;
};

// Obtener ratings promedio para múltiples películas usando tmdbIds
// Si no hay rating en CSV, usa el rating de TMDB normalizado (0-10 -> 0-5)
export const getMoviesRatingsByTMDBIds = async (
  tmdbIds: number[],
  tmdbRatings?: Map<number, number> // Map<tmdbId, rating> de TMDB
): Promise<Map<number, number>> => {
  const mapping = await loadTMDBToMovieIdMapping();
  const averages = await loadRatingsFromCSV();
  const result = new Map<number, number>();
  
  console.log('Mapeo tmdbId -> movieId:', mapping.size, 'entradas');
  console.log('Ratings promedio:', averages.size, 'entradas');
  console.log('tmdbIds a buscar:', tmdbIds.length);
  
  tmdbIds.forEach(tmdbId => {
    const csvMovieId = mapping.get(tmdbId);
    let rating: number | undefined;
    
    // Primero intentar obtener rating del CSV
    if (csvMovieId) {
      rating = averages.get(csvMovieId);
      if (rating !== undefined) {
        result.set(tmdbId, rating);
        console.log(`Encontrado rating CSV para tmdbId ${tmdbId} (movieId ${csvMovieId}): ${rating}`);
      }
    }
    
    // Si no hay rating en CSV, usar rating de TMDB normalizado
    if (rating === undefined && tmdbRatings) {
      const tmdbRating = tmdbRatings.get(tmdbId);
      if (tmdbRating !== undefined && tmdbRating > 0) {
        const normalizedRating = normalizeTMDBRating(tmdbRating);
        result.set(tmdbId, normalizedRating);
        console.log(`Usando rating TMDB normalizado para tmdbId ${tmdbId}: ${tmdbRating} -> ${normalizedRating.toFixed(2)}`);
      }
    }
  });
  
  console.log('Ratings encontrados:', result.size);
  return result;
};

