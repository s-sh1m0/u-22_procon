import { Routes, Route } from 'react-router-dom'
import Login from './pages/Login'
import Home from './pages/Home'
import Analysis from './pages/Analysis'
import { RequireAuth } from './components/auth/RequireAuth'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<RequireAuth><Home /></RequireAuth>} />
      <Route path="/analysis/:jobId" element={<RequireAuth><Analysis /></RequireAuth>} />
    </Routes>
  )
}
