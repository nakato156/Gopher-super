import { Movie } from './tmdb';
import { API_CONFIG } from '../config';

const API_URL = API_CONFIG.BASE_URL;

const getAuthHeaders = () => {
    const token = localStorage.getItem('authToken');
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
    };
};

interface RecommendationResponse {
    recommendations: {
        movieId: number;
        title: string;
        genres: string[];
        rating: number;
        score: number;
    }[];
}

interface PopularResponse {
    movies: {
        movieId: number;
        title: string;
        genres: string[];
        rating: number;
    }[];
}

export const getRecommendations = async (topN: number = 10): Promise<Movie[]> => {
    try {
        const response = await fetch(`${API_URL}/api/recomend`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ top_n: topN })
        });

        if (!response.ok) {
            throw new Error('Failed to fetch recommendations');
        }

        const data: RecommendationResponse = await response.json();
        console.log({ data });
        const r = data.recommendations.map(r => ({
            movieId: r.movieId,
            title: r.title,
            genres: Array.isArray(r.genres) ? r.genres.join('|') : r.genres,
            rating: r.rating,
            csvRating: r.rating,
            overview: `Score: ${r.score.toFixed(4)}`
        }));
        console.log({ r });
        return r;
    } catch (error) {
        console.error('Error fetching recommendations:', error);
        return [];
    }
};

export const getPopularMovies = async (topN: number = 10): Promise<Movie[]> => {
    try {
        const response = await fetch(`${API_URL}/api/recomend/popular?top_n=${topN}`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            console.log({ res: response.status })
            throw new Error('Failed to fetch popular movies ya fue');
        }

        const data: PopularResponse = await response.json();

        const a = data.movies.map(m => ({
            movieId: m.movieId,
            title: m.title,
            genres: Array.isArray(m.genres) ? m.genres.join('|') : m.genres,
            rating: m.rating,
            csvRating: m.rating
        }));
        console.log({ a });
        return a;
    } catch (error) {
        console.error('Error fetching popular movies:', error);
        return [];
    }
};
