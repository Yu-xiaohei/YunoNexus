import { create } from 'zustand'

interface AuthState {
  token: string | null
  user: Record<string, unknown> | null
  setToken: (token: string) => void
  setUser: (user: Record<string, unknown>) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: null,
  setToken: (token) => {
    localStorage.setItem('token', token)
    set({ token })
  },
  setUser: (user) => set({ user }),
  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null })
    window.location.href = '/login'
  },
}))
