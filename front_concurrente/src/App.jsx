import { useState, useEffect } from 'react'
import Login from './movie_recomend/login/login'
import Dashboard from './movie_recomend/user_dashboard/dashboard'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

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
  }

  if (isAuthenticated) {
    return <Dashboard onLogout={handleLogout} />
  }

  return <Login onLoginSuccess={handleLoginSuccess} />
}

export default App
