import { Movie } from './tmdb';
import { API_CONFIG } from '../config';

const API_URL = API_CONFIG.BASE_URL;

export interface GenreStat {
    genre: string;
    count: number;
}

const getAuthHeaders = () => {
    const token = localStorage.getItem('authToken');
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
    };
};

export const getUserMoviesSeen = async (): Promise<Movie[]> => {
    try {
        const response = await fetch(`${API_URL}/api/user/movies`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch user movies');
        }

        const data = await response.json();
        // Map backend movie format to frontend Movie interface if necessary
        // Backend: { movieId, title, genres }
        // Frontend Movie: { movieId, title, genres, ... }
        // The backend returns genres as array of strings, frontend expects string joined by '|'?
        // Let's check frontend Movie interface in tmdb.ts
        return data.map((m: any) => ({
            ...m,
            genres: Array.isArray(m.genres) ? m.genres.join('|') : m.genres,
            csvRating: m.rating // Map backend rating to frontend csvRating for display
        }));
    } catch (error) {
        console.error('Error fetching user movies:', error);
        return [];
    }
};

export const getUserTopGenres = async (): Promise<GenreStat[]> => {
    try {
        const response = await fetch(`${API_URL}/api/user/stats/genres`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch top genres');
        }

        return await response.json();
    } catch (error) {
        console.error('Error fetching top genres:', error);
        return [];
    }
};
