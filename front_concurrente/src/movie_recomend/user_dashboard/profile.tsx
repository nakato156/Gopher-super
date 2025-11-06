import React, { useState, useEffect, useMemo } from 'react';
import { Film, User, Calendar, ArrowLeft, Lock, UserCircle, Mail } from 'lucide-react';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Movie } from '../services/tmdb';
import { getUserMovieList, getUserAccountDate } from '../services/userList';
import MovieCard from '../components/MovieCard';

interface ProfileProps {
  onBack?: () => void;
}

const Profile: React.FC<ProfileProps> = ({ onBack }) => {
  const [userName, setUserName] = useState<string>('Usuario');
  const [accountDate, setAccountDate] = useState<Date>(new Date());
  const [userMovies, setUserMovies] = useState<Movie[]>([]);

  // Función para actualizar la lista de películas
  const updateMovieList = () => {
    const movies = getUserMovieList();
    setUserMovies(movies);
  };

  // Obtener datos del usuario
  useEffect(() => {
    const userEmail = localStorage.getItem('userEmail');
    if (userEmail) {
      const name = userEmail.split('@')[0];
      setUserName(name.charAt(0).toUpperCase() + name.slice(1));
    }
    
    const date = getUserAccountDate();
    setAccountDate(date);
    
    updateMovieList();

    // Escuchar cambios en localStorage (cuando se agregan/quitan películas)
    const handleStorageChange = () => {
      updateMovieList();
    };

    window.addEventListener('storage', handleStorageChange);
    
    // También escuchar eventos personalizados si se usan
    const interval = setInterval(() => {
      updateMovieList();
    }, 1000); // Actualizar cada segundo para detectar cambios

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      clearInterval(interval);
    };
  }, []);

  // Calcular datos para el gráfico de pie (películas por género)
  const genrePieData = useMemo(() => {
    const genreCount: { [key: string]: number } = {};
    
    userMovies.forEach(movie => {
      const genres = movie.genres.split('|').filter(g => g.trim() !== '');
      if (genres.length === 0) {
        genreCount['Unknown'] = (genreCount['Unknown'] || 0) + 1;
      } else {
        genres.forEach(genre => {
          genreCount[genre] = (genreCount[genre] || 0) + 1;
        });
      }
    });

    return Object.entries(genreCount).map(([name, value]) => ({
      name,
      value,
    }));
  }, [userMovies]);

  // Calcular datos para el gráfico de columnas (rating promedio por género)
  const genreRatingData = useMemo(() => {
    const genreRatings: { [key: string]: { total: number; count: number } } = {};
    
    userMovies.forEach(movie => {
      const genres = movie.genres.split('|').filter(g => g.trim() !== '');
      const rating = movie.csvRating || movie.rating ? (movie.csvRating || (movie.rating! / 2)) : 0;
      
      if (genres.length === 0) {
        if (!genreRatings['Unknown']) {
          genreRatings['Unknown'] = { total: 0, count: 0 };
        }
        genreRatings['Unknown'].total += rating;
        genreRatings['Unknown'].count += 1;
      } else {
        genres.forEach(genre => {
          if (!genreRatings[genre]) {
            genreRatings[genre] = { total: 0, count: 0 };
          }
          genreRatings[genre].total += rating;
          genreRatings[genre].count += 1;
        });
      }
    });

    return Object.entries(genreRatings).map(([name, data]) => ({
      name,
      rating: data.count > 0 ? data.total / data.count : 0,
    })).sort((a, b) => b.rating - a.rating);
  }, [userMovies]);

  // Colores para el gráfico de pie
  const COLORS = [
    '#667eea',
    '#764ba2',
    '#f093fb',
    '#4facfe',
    '#00f2fe',
    '#43e97b',
    '#fa709a',
    '#fee140',
    '#30cfd0',
    '#a8edea',
    '#fed6e3',
    '#a8caba',
  ];

  const formatDate = (date: Date): string => {
    return new Intl.DateTimeFormat('es-ES', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    }).format(date);
  };

  const styles = {
    container: {
      height: '100vh',
      width: '100%',
      background: '#0f0f0f',
      color: '#ffffff',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
      display: 'flex',
      padding: '0',
      margin: '0',
      overflow: 'hidden' as const,
    },
    sidebar: {
      width: '280px',
      height: '100vh',
      background: 'linear-gradient(180deg, #667eea 0%, #764ba2 100%)',
      padding: '40px 24px',
      display: 'flex',
      flexDirection: 'column' as const,
      alignItems: 'center',
      boxShadow: '4px 0 16px rgba(0, 0, 0, 0.3)',
      position: 'relative' as const,
      overflowY: 'auto' as const,
    },
    topSection: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: '16px',
      width: '100%',
      marginTop: '0',
      marginBottom: '40px',
    },
    backButton: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '10px',
      background: 'rgba(255, 255, 255, 0.2)',
      border: '1px solid rgba(255, 255, 255, 0.3)',
      borderRadius: '8px',
      color: '#ffffff',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      width: '40px',
      height: '40px',
      flexShrink: 0,
    },
    appName: {
      fontSize: '28px',
      fontWeight: 'bold' as const,
      color: '#ffffff',
      letterSpacing: '-0.5px',
      textAlign: 'right' as const,
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      marginLeft: 'auto',
    },
    userSection: {
      display: 'flex',
      flexDirection: 'column' as const,
      alignItems: 'center',
      gap: '20px',
      marginBottom: '60px',
    },
    userIcon: {
      width: '80px',
      height: '80px',
      borderRadius: '50%',
      background: 'rgba(255, 255, 255, 0.2)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      border: '3px solid rgba(255, 255, 255, 0.3)',
    },
    userName: {
      fontSize: '20px',
      fontWeight: '600' as const,
      color: '#ffffff',
      textAlign: 'center' as const,
    },
    accountDate: {
      fontSize: '14px',
      color: 'rgba(255, 255, 255, 0.8)',
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      textAlign: 'center' as const,
    },
    settingsSection: {
      display: 'flex',
      flexDirection: 'column' as const,
      gap: '12px',
      width: '100%',
      marginTop: '0',
      paddingTop: '220px',
      marginBottom: '30px',
    },
    settingsButton: {
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      padding: '12px 16px',
      background: 'rgba(255, 255, 255, 0.15)',
      border: '1px solid rgba(255, 255, 255, 0.2)',
      borderRadius: '8px',
      color: '#ffffff',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      fontSize: '14px',
      fontWeight: '500' as const,
      width: '100%',
      textAlign: 'left' as const,
    },
    content: {
      flex: 1,
      padding: '40px',
      overflowY: 'auto' as const,
      height: '100vh',
      boxSizing: 'border-box' as const,
    },
    chartsContainer: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '32px',
      marginBottom: '40px',
    },
    chartCard: {
      background: '#1a1825',
      borderRadius: '12px',
      padding: '24px',
      border: '1px solid #2a2640',
    },
    chartTitle: {
      fontSize: '18px',
      fontWeight: '600' as const,
      color: '#ffffff',
      marginBottom: '20px',
    },
    moviesList: {
      background: '#1a1825',
      borderRadius: '12px',
      padding: '24px',
      border: '1px solid #2a2640',
    },
    moviesListTitle: {
      fontSize: '20px',
      fontWeight: '600' as const,
      color: '#ffffff',
      marginBottom: '24px',
    },
    moviesGrid: {
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

  const handleSettingsClick = (action: string) => {
    // Aquí puedes implementar la lógica para cada acción
    console.log(`Acción: ${action}`);
    // Por ejemplo, abrir modales o navegar a otras páginas
  };

  return (
    <div style={styles.container}>
      {/* Sidebar */}
      <div style={styles.sidebar}>
        {/* Sección superior con botón y título */}
        <div style={styles.topSection}>
          {/* Botón para regresar al dashboard */}
          {onBack && (
            <button
              style={styles.backButton}
              onClick={onBack}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(255, 255, 255, 0.3)';
                e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.5)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(255, 255, 255, 0.2)';
                e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.3)';
              }}
              title="Regresar al dashboard"
            >
              <ArrowLeft size={20} />
            </button>
          )}

          {/* Nombre de la app */}
          <div style={styles.appName}>
            <Film size={32} />
            <span>MOVIEFLIX</span>
          </div>
        </div>
        
        {/* Sección de usuario (icono y nombre) */}
        <div style={styles.userSection}>
          <div style={styles.userIcon}>
            <User size={40} />
          </div>
          <div style={styles.userName}>{userName}</div>
          <div style={styles.accountDate}>
            <Calendar size={16} />
            <span>{formatDate(accountDate)}</span>
          </div>
        </div>

        {/* Botones de configuración abajo */}
        <div style={styles.settingsSection}>
          <button
            style={styles.settingsButton}
            onClick={() => handleSettingsClick('changePassword')}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.25)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.4)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.15)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.2)';
            }}
          >
            <Lock size={18} />
            <span>Cambiar contraseña</span>
          </button>
          <button
            style={styles.settingsButton}
            onClick={() => handleSettingsClick('changeUsername')}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.25)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.4)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.15)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.2)';
            }}
          >
            <UserCircle size={18} />
            <span>Cambiar nombre de usuario</span>
          </button>
          <button
            style={styles.settingsButton}
            onClick={() => handleSettingsClick('changeEmail')}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.25)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.4)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.15)';
              e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.2)';
            }}
          >
            <Mail size={18} />
            <span>Cambiar correo de usuario</span>
          </button>
        </div>
      </div>

      {/* Content */}
      <div style={styles.content}>
        {/* Charts */}
        <div style={styles.chartsContainer}>
          {/* Pie Chart */}
          <div style={styles.chartCard}>
            <div style={styles.chartTitle}>Películas por Género</div>
            {genrePieData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={genrePieData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={(props: any) => `${props.name}: ${(props.percent * 100).toFixed(0)}%`}
                    outerRadius={100}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {genrePieData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ 
                      background: '#1a1825', 
                      border: '1px solid #2a2640',
                      borderRadius: '8px',
                      color: '#ffffff'
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div style={styles.emptyState}>No hay datos para mostrar</div>
            )}
          </div>

          {/* Bar Chart */}
          <div style={styles.chartCard}>
            <div style={styles.chartTitle}>Rating Promedio por Género</div>
            {genreRatingData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={genreRatingData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#2a2640" />
                  <XAxis 
                    dataKey="name" 
                    stroke="#8b87a0"
                    tick={{ fill: '#8b87a0', fontSize: 12 }}
                    angle={-45}
                    textAnchor="end"
                    height={80}
                  />
                  <YAxis 
                    stroke="#8b87a0"
                    tick={{ fill: '#8b87a0', fontSize: 12 }}
                    domain={[0, 5]}
                  />
                  <Tooltip 
                    contentStyle={{ 
                      background: '#1a1825', 
                      border: '1px solid #2a2640',
                      borderRadius: '8px',
                      color: '#ffffff'
                    }}
                    formatter={(value: number) => value.toFixed(2)}
                  />
                  <Legend 
                    wrapperStyle={{ color: '#ffffff' }}
                  />
                  <Bar 
                    dataKey="rating" 
                    fill="#667eea"
                    radius={[8, 8, 0, 0]}
                  />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div style={styles.emptyState}>No hay datos para mostrar</div>
            )}
          </div>
        </div>

        {/* Movies List */}
        <div style={styles.moviesList}>
          <div style={styles.moviesListTitle}>Mi Lista de Películas</div>
          {userMovies.length > 0 ? (
            <div style={styles.moviesGrid}>
              {userMovies.map((movie) => (
                <MovieCard 
                  key={movie.movieId} 
                  movie={movie} 
                  size="small"
                />
              ))}
            </div>
          ) : (
            <div style={styles.emptyState}>
              No tienes películas en tu lista aún. Agrega películas desde el dashboard principal.
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Profile;

