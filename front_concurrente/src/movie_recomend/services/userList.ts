// Servicio para gestionar la lista de películas del usuario
import { Movie } from './tmdb';

const USER_LIST_KEY = 'userMovieList';
const USER_ACCOUNT_DATE_KEY = 'userAccountDate';

// Obtener la lista de películas del usuario
export const getUserMovieList = (): Movie[] => {
  try {
    const stored = localStorage.getItem(USER_LIST_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
    return [];
  } catch (error) {
    console.error('Error loading user movie list:', error);
    return [];
  }
};

// Agregar una película a la lista del usuario
export const addMovieToList = (movie: Movie): void => {
  try {
    const currentList = getUserMovieList();
    // Verificar si la película ya está en la lista
    const exists = currentList.some(m => m.movieId === movie.movieId);
    if (!exists) {
      const updatedList = [...currentList, movie];
      localStorage.setItem(USER_LIST_KEY, JSON.stringify(updatedList));
    }
  } catch (error) {
    console.error('Error adding movie to list:', error);
  }
};

// Eliminar una película de la lista del usuario
export const removeMovieFromList = (movieId: number): void => {
  try {
    const currentList = getUserMovieList();
    const updatedList = currentList.filter(m => m.movieId !== movieId);
    localStorage.setItem(USER_LIST_KEY, JSON.stringify(updatedList));
  } catch (error) {
    console.error('Error removing movie from list:', error);
  }
};

// Verificar si una película está en la lista
export const isMovieInList = (movieId: number): boolean => {
  const list = getUserMovieList();
  return list.some(m => m.movieId === movieId);
};

// Obtener la fecha de creación de la cuenta del usuario
export const getUserAccountDate = (): Date => {
  try {
    const stored = localStorage.getItem(USER_ACCOUNT_DATE_KEY);
    if (stored) {
      return new Date(stored);
    }
    // Si no existe, crear una fecha por defecto (fecha actual)
    const now = new Date();
    localStorage.setItem(USER_ACCOUNT_DATE_KEY, now.toISOString());
    return now;
  } catch (error) {
    console.error('Error loading user account date:', error);
    return new Date();
  }
};

// Establecer la fecha de creación de la cuenta (solo si no existe)
export const setUserAccountDate = (date: Date): void => {
  try {
    const existing = localStorage.getItem(USER_ACCOUNT_DATE_KEY);
    if (!existing) {
      localStorage.setItem(USER_ACCOUNT_DATE_KEY, date.toISOString());
    }
  } catch (error) {
    console.error('Error setting user account date:', error);
  }
};

