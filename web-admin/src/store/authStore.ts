import { create } from 'zustand'

interface AuthState {
  token: string | null
  user: Record<string, unknown> | null
  setToken: (token: string) => void
  setUser: (user: Record<string, unknown>) => void
  logout: () => void
  loadUser: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  setToken: (token) => {
    localStorage.setItem('token', token)
    set({ token })
  },
  setUser: (user) => {
    localStorage.setItem('user', JSON.stringify(user))
    set({ user })
  },
  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null })
    window.location.href = '/login'
  },
  loadUser: () => {
    const user = JSON.parse(localStorage.getItem('user') || 'null')
    set({ user })
  },
}))
