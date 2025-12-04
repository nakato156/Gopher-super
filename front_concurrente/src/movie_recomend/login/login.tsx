import React, { useState, FormEvent, ChangeEvent, useEffect } from 'react';
import { Eye, EyeOff, Film, ChevronRight } from 'lucide-react';
import intelestelarImg from '../../assets/movie_imgs/intelestelar.jpg';
import herImg from '../../assets/movie_imgs/her.jpg';
import bladerunnerImg from '../../assets/movie_imgs/bladerrunnner.jpg';

interface FormData {
  email: string;
  password: string;
  rememberMe: boolean;
}

interface FormErrors {
  email?: string;
  password?: string;
}

interface LoginProps {
  onLoginSuccess?: () => void;
}

const Login: React.FC<LoginProps> = ({ onLoginSuccess }) => {
  const [formData, setFormData] = useState<FormData>({
    email: '',
    password: '',
    rememberMe: false
  });

  const [showPassword, setShowPassword] = useState<boolean>(false);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [currentMovieIndex, setCurrentMovieIndex] = useState<number>(0);

  // Imágenes locales de películas
  const moviePosters: string[] = [
    intelestelarImg,
    herImg,
    bladerunnerImg,
  ];

  const handleChange = (e: ChangeEvent<HTMLInputElement>): void => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));

    if (errors[name as keyof FormErrors]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }));
    }
  };

  const validateForm = (): FormErrors => {
    const newErrors: FormErrors = {};

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    return newErrors;
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    const newErrors = validateForm();

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setIsLoading(true);

    try {
      // Aquí iría tu llamada a la API
      await new Promise(resolve => setTimeout(resolve, 2000));
      console.log('Login data:', formData);
      // Guardar el email del usuario en localStorage
      localStorage.setItem('userEmail', formData.email);
      // Manejar respuesta exitosa
      if (onLoginSuccess) {
        onLoginSuccess();
      }
    } catch (error) {
      console.error('Login error:', error);
      // Manejar error
    } finally {
      setIsLoading(false);
    }
  };

  const handleSocialLogin = (provider: 'google' | 'apple'): void => {
    console.log(`Login with ${provider}`);
    // Implementar lógica de login social
  };

  // Efecto para cambiar automáticamente las imágenes cada 5 segundos
  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentMovieIndex((prev) => (prev + 1) % moviePosters.length);
    }, 5000);
    return () => clearInterval(interval);
  }, [moviePosters.length]);

  const styles = {
    container: {
      minHeight: '100vh',
      height: '100vh',
      width: '100%',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '0',
      margin: '0',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
      boxSizing: 'border-box' as const,
    },
    card: {
      width: '100%',
      height: '100%',
      maxWidth: '100%',
      display: 'flex',
      background: '#1a1825',
      borderRadius: '0',
      overflow: 'hidden',
      boxShadow: 'none',
    },
    leftSection: {
      flex: 1,
      background: 'transparent',
      padding: '0',
      display: 'flex',
      flexDirection: 'column' as const,
      position: 'relative' as const,
      justifyContent: 'space-between',
      overflow: 'hidden',
    },
    movieGallery: {
      position: 'absolute' as const,
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '0',
      width: '100%',
      height: '100%',
    },
    moviePoster: {
      width: '100%',
      height: '100%',
      objectFit: 'cover' as const,
      borderRadius: '0',
      boxShadow: 'none',
    },
    movieContent: {
      position: 'relative' as const,
      zIndex: 2,
      width: '100%',
      padding: '40px',
      display: 'flex',
      flexDirection: 'column' as const,
      justifyContent: 'space-between',
      height: '100%',
    },
    movieDots: {
      display: 'flex',
      gap: '8px',
      justifyContent: 'center',
      marginTop: '20px',
    },
    movieDot: {
      width: '8px',
      height: '8px',
      borderRadius: '50%',
      background: 'rgba(255, 255, 255, 0.4)',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
    },
    activeMovieDot: {
      width: '32px',
      borderRadius: '4px',
      background: 'white',
    },
    logo: {
      display: 'flex',
      alignItems: 'center',
      gap: '10px',
      color: 'white',
    },
    logoText: {
      fontSize: '24px',
      fontWeight: 'bold',
      letterSpacing: '-0.5px',
    },
    tagline: {
      fontSize: '42px',
      fontWeight: '300',
      color: 'white',
      lineHeight: '1.2',
      letterSpacing: '-1px',
      marginBottom: '30px',
    },
    dots: {
      display: 'flex',
      gap: '8px',
    },
    dot: {
      width: '8px',
      height: '8px',
      borderRadius: '50%',
      background: 'rgba(255, 255, 255, 0.3)',
    },
    activeDot: {
      width: '32px',
      borderRadius: '4px',
      background: 'white',
    },
    rightSection: {
      flex: 1.2,
      background: '#1a1825',
      padding: '40px',
      display: 'flex',
      flexDirection: 'column' as const,
      position: 'relative' as const,
    },
    backButton: {
      position: 'absolute' as const,
      top: '40px',
      right: '40px',
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      background: 'transparent',
      border: 'none',
      color: '#8b87a0',
      fontSize: '14px',
      cursor: 'pointer',
    },
    formContainer: {
      display: 'flex',
      flexDirection: 'column' as const,
      justifyContent: 'center',
      height: '100%',
      maxWidth: '400px',
      margin: '0 auto',
      width: '100%',
    },
    title: {
      fontSize: '60px',
      color: 'white',
      marginBottom: '12px',
      fontWeight: '600',
    },
    subtitle: {
      color: '#8b87a0',
      fontSize: '14px',
      marginBottom: '32px',
    },
    link: {
      color: '#667eea',
      background: 'none',
      border: 'none',
      cursor: 'pointer',
      textDecoration: 'underline',
      marginLeft: '4px',
    },
    form: {
      display: 'flex',
      flexDirection: 'column' as const,
      gap: '20px',
    },
    formGroup: {
      position: 'relative' as const,
      width: '100%',
      boxSizing: 'border-box' as const,
    },
    input: {
      width: '100%',
      padding: '14px 16px',
      background: '#262335',
      border: '2px solid transparent',
      borderRadius: '12px',
      color: 'white',
      fontSize: '14px',
      outline: 'none',
      transition: 'all 0.3s ease',
      boxSizing: 'border-box' as const,
    },
    inputError: {
      borderColor: '#ef4444',
    },
    passwordWrapper: {
      position: 'relative' as const,
      width: '100%',
      boxSizing: 'border-box' as const,
    },
    passwordToggle: {
      position: 'absolute' as const,
      right: '16px',
      top: '50%',
      transform: 'translateY(-50%)',
      background: 'none',
      border: 'none',
      color: '#6b6878',
      cursor: 'pointer',
      padding: '0',
      display: 'flex',
      alignItems: 'center',
    },
    errorMessage: {
      color: '#ef4444',
      fontSize: '12px',
      marginTop: '6px',
      display: 'block',
    },
    formOptions: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
    },
    checkbox: {
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
      color: '#8b87a0',
      fontSize: '14px',
    },
    forgotLink: {
      background: 'none',
      border: 'none',
      color: '#667eea',
      fontSize: '14px',
      cursor: 'pointer',
    },
    submitButton: {
      width: '100%',
      padding: '14px',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      border: 'none',
      borderRadius: '12px',
      color: 'white',
      fontSize: '16px',
      fontWeight: '500',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
      marginTop: '8px',
    },
    submitButtonDisabled: {
      opacity: 0.5,
      cursor: 'not-allowed',
    },
    divider: {
      textAlign: 'center' as const,
      margin: '24px 0',
      position: 'relative' as const,
      color: '#6b6878',
      fontSize: '14px',
    },
    dividerLine: {
      position: 'absolute' as const,
      left: '0',
      top: '50%',
      width: '100%',
      height: '1px',
      background: '#2a2640',
      zIndex: 0,
    },
    dividerText: {
      background: '#1a1825',
      padding: '0 16px',
      position: 'relative' as const,
      zIndex: 1,
    },
    socialButtons: {
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: '12px',
    },
    socialButton: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '10px',
      padding: '12px',
      background: 'transparent',
      border: '2px solid #2a2640',
      borderRadius: '12px',
      color: '#8b87a0',
      fontSize: '14px',
      fontWeight: '500',
      cursor: 'pointer',
      transition: 'all 0.3s ease',
    },
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        {/* Left Section */}
        <div style={styles.leftSection}>
          {/* Galería de imágenes de películas */}
          <div style={styles.movieGallery}>
            <img
              src={moviePosters[currentMovieIndex]}
              alt={`Movie poster ${currentMovieIndex + 1}`}
              style={styles.moviePoster}
              onError={(e) => {
                // Fallback si la imagen no carga
                (e.target as HTMLImageElement).src = 'https://via.placeholder.com/500x750/667eea/ffffff?text=Movie+Poster';
              }}
            />
          </div>

          {/* Contenido sobre las imágenes */}
          <div style={styles.movieContent}>
            <div style={styles.logo}>
              <Film size={32} />
              <span style={styles.logoText}>GoFlix</span>
            </div>

            <div style={{ marginTop: 'auto' }}>
              <div style={styles.movieDots}>
                {moviePosters.map((_, index) => (
                  <span
                    key={index}
                    style={{
                      ...styles.movieDot,
                      ...(index === currentMovieIndex ? styles.activeMovieDot : {})
                    }}
                    onClick={() => setCurrentMovieIndex(index)}
                  ></span>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Right Section */}
        <div style={styles.rightSection}>
          <button style={styles.backButton}>
            Back to website
            <ChevronRight size={16} />
          </button>

          <div style={styles.formContainer}>
            <h2 style={styles.title}>Welcome back</h2>
            <p style={styles.subtitle}>
              New to GoFlix?
              <button style={styles.link}>
                Create an account
              </button>
            </p>

            <form onSubmit={handleSubmit} style={styles.form}>
              <div style={styles.formGroup}>
                <input
                  type="email"
                  name="email"
                  placeholder="Email"
                  value={formData.email}
                  onChange={handleChange}
                  style={{
                    ...styles.input,
                    ...(errors.email ? styles.inputError : {})
                  }}
                  autoComplete="email"
                />
                {errors.email && (
                  <span style={styles.errorMessage}>{errors.email}</span>
                )}
              </div>

              <div style={styles.formGroup}>
                <div style={styles.passwordWrapper}>
                  <input
                    type={showPassword ? 'text' : 'password'}
                    name="password"
                    placeholder="Enter your password"
                    value={formData.password}
                    onChange={handleChange}
                    style={{
                      ...styles.input,
                      paddingRight: '50px',
                      ...(errors.password ? styles.inputError : {})
                    }}
                    autoComplete="current-password"
                  />
                  <button
                    type="button"
                    style={styles.passwordToggle}
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                  </button>
                </div>
                {errors.password && (
                  <span style={styles.errorMessage}>{errors.password}</span>
                )}
              </div>

              <div style={styles.formOptions}>
                <label style={styles.checkbox}>
                  <input
                    type="checkbox"
                    name="rememberMe"
                    checked={formData.rememberMe}
                    onChange={handleChange}
                  />
                  <span>Remember me</span>
                </label>
                <button type="button" style={styles.forgotLink}>
                  Forgot password?
                </button>
              </div>

              <button
                type="submit"
                style={{
                  ...styles.submitButton,
                  ...(isLoading ? styles.submitButtonDisabled : {})
                }}
                disabled={isLoading}
              >
                {isLoading ? 'Signing in...' : 'Sign in'}
              </button>
            </form>

            <div style={styles.divider}>
              <div style={styles.dividerLine}></div>
              <span style={styles.dividerText}>Or sign in with</span>
            </div>

            <div style={styles.socialButtons}>
              <button
                style={styles.socialButton}
                onClick={() => handleSocialLogin('google')}
              >
                <svg width="20" height="20" viewBox="0 0 24 24">
                  <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                  <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                  <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                  <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                </svg>
                Google
              </button>

              <button
                style={styles.socialButton}
                onClick={() => handleSocialLogin('apple')}
              >
                <svg width="20" height="20" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09l.01-.01zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z" />
                </svg>
                Apple
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;