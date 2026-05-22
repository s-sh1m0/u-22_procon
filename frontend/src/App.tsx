import { Routes, Route } from 'react-router-dom'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Home from './pages/Home'
import Analysis from './pages/Analysis'
import { RequireAuth } from './components/auth/RequireAuth'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route
        path="/analyze"
        element={
          <RequireAuth>
            <Home />
          </RequireAuth>
        }
      />
      <Route
        path="/analysis/:jobId"
        element={
          <RequireAuth>
            <Analysis />
          </RequireAuth>
        }
      />
    </Routes>
  )
}
