import React, { useState, FormEvent, useEffect } from 'react';
import { Search } from 'lucide-react';

interface SearchBarProps {
  placeholder?: string;
  onSearch?: (query: string) => void;
  onQueryChange?: (query: string) => void;
}

const SearchBar: React.FC<SearchBarProps> = ({ 
  placeholder = 'Buscar películas...',
  onSearch,
  onQueryChange
}) => {
  const [searchQuery, setSearchQuery] = useState('');

  // Debounce para búsqueda en tiempo real
  useEffect(() => {
    if (!onSearch) return;

    const timeoutId = setTimeout(() => {
      if (searchQuery.trim().length >= 2 || searchQuery.trim().length === 0) {
        onSearch(searchQuery.trim());
      }
    }, 300); // Esperar 300ms después de que el usuario deje de escribir

    return () => clearTimeout(timeoutId);
  }, [searchQuery, onSearch]);

  const styles = {
    searchContainer: {
      position: 'relative' as const,
      width: '100%',
      maxWidth: '600px',
      margin: '0 auto',
    },
    searchInput: {
      width: '100%',
      padding: '12px 16px 12px 48px',
      background: '#262335',
      border: '2px solid #2a2640',
      borderRadius: '12px',
      color: '#ffffff',
      fontSize: '14px',
      outline: 'none',
      transition: 'all 0.3s ease',
    },
    searchIcon: {
      position: 'absolute' as const,
      left: '16px',
      top: '50%',
      transform: 'translateY(-50%)',
      color: '#8b87a0',
      pointerEvents: 'none' as const,
    },
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (onSearch && searchQuery.trim()) {
      onSearch(searchQuery.trim());
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setSearchQuery(value);
    // Notificar cambio de query para resetear el estado de cierre manual
    if (onQueryChange) {
      onQueryChange(value);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={styles.searchContainer}>
      <Search 
        size={20} 
        style={styles.searchIcon}
      />
      <input
        type="text"
        value={searchQuery}
        onChange={handleChange}
        placeholder={placeholder}
        style={styles.searchInput}
        onFocus={(e) => {
          e.currentTarget.style.borderColor = '#667eea';
        }}
        onBlur={(e) => {
          e.currentTarget.style.borderColor = '#2a2640';
        }}
      />
    </form>
  );
};

export default SearchBar;

