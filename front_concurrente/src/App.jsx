import { useState, useEffect } from 'react'
import Login from './movie_recomend/login/login'
import Dashboard from './movie_recomend/user_dashboard/dashboard'
import Profile from './movie_recomend/user_dashboard/profile'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [currentView, setCurrentView] = useState('dashboard') // 'dashboard' o 'profile'

  // Verificar si hay una sesión guardada al cargar la app
  useEffect(() => {
    const session = localStorage.getItem('isAuthenticated')
    if (session === 'true') {
      setIsAuthenticated(true)
    }
  }, [])

  const handleLoginSuccess = () => {
    localStorage.setItem('isAuthenticated', 'true')
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    localStorage.removeItem('isAuthenticated')
    localStorage.removeItem('userEmail')
    setIsAuthenticated(false)
    setCurrentView('dashboard')
  }

  if (isAuthenticated) {
    if (currentView === 'profile') {
      return <Profile onBack={() => setCurrentView('dashboard')} />
    }
    return <Dashboard onLogout={handleLogout} onNavigateToProfile={() => setCurrentView('profile')} />
  }

  return <Login onLoginSuccess={handleLoginSuccess} />
}

export default App
