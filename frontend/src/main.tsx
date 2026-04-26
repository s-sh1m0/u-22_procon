import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
// TODO: import { QueryClientProvider } from '@tanstack/react-query'
// TODO: import { queryClient } from './lib/queryClient'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      {/* TODO: <QueryClientProvider client={queryClient}> */}
      <App />
      {/* TODO: </QueryClientProvider> */}
    </BrowserRouter>
  </StrictMode>,
)
